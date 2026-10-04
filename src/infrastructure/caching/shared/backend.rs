use std::{future::Future, sync::Arc, time::Duration};

use redis::{
    ConnectionAddr, IntoConnectionInfo, RedisConnectionInfo,
    aio::{ConnectionManager, ConnectionManagerConfig},
};
use serde::{Serialize, de::DeserializeOwned};

use super::{
    error,
    query::{self, PayloadKind},
    scripts, validation,
};
use crate::domain::{
    contracts::{caching::KeyBuilder, utility::Logger},
    models::{
        AppContext, AppEnv, CacheEntryKey, CacheFamily, CacheQuery, CacheRead, CacheRevision,
        CacheStamp, CacheStoreOutcome, CachingError,
    },
};

/// Shared standalone Redis connection and private execution helpers.
#[derive(Clone)]
pub struct RedisBackend {
    connection: ConnectionManager,
    key_builder: Arc<dyn KeyBuilder>,
    logger: Arc<dyn Logger>,
    operation_timeout: Duration,
}

impl RedisBackend {
    pub async fn connect(
        context: &AppContext,
        env: &AppEnv,
        key_builder: Arc<dyn KeyBuilder>,
        logger: Arc<dyn Logger>,
    ) -> Result<Self, CachingError> {
        let result = async {
            error::context(context)?;

            if env.redis_host.is_empty()
                || env.redis_port == 0
                || env.redis_connect_timeout.is_zero()
                || env.redis_operation_timeout.is_zero()
            {
                return Err(CachingError::BadArgs);
            }

            let mut redis = RedisConnectionInfo::default().set_db(i64::from(env.redis_database));
            if !env.redis_password.is_empty() {
                redis = redis.set_password(&env.redis_password);
            }

            let client = redis::Client::open(
                ConnectionAddr::Tcp(env.redis_host.clone(), env.redis_port)
                    .into_connection_info()
                    .map_err(error::redis)?
                    .set_redis_settings(redis),
            )
            .map_err(error::redis)?;

            // Cache callers fall back rather than waiting through handshake retries.
            // The manager still reconnects automatically after connection loss.
            let config = ConnectionManagerConfig::new()
                .set_number_of_retries(0)
                .set_connection_timeout(Some(env.redis_connect_timeout))
                .set_response_timeout(Some(env.redis_operation_timeout));

            let connection = Self::bounded(env.redis_connect_timeout, async {
                ConnectionManager::new_with_config(client, config)
                    .await
                    .map_err(error::redis)
            })
            .await?;

            Ok(Self {
                connection,
                key_builder,
                logger: logger.clone(),
                operation_timeout: env.redis_operation_timeout,
            })
        }
        .await;

        error::finish(logger.as_ref(), context, "caching/redis/connect", result)
    }

    async fn bounded<T>(
        duration: Duration,
        operation: impl Future<Output = Result<T, CachingError>>,
    ) -> Result<T, CachingError> {
        tokio::time::timeout(duration, operation)
            .await
            .map_err(|source| CachingError::Timeout {
                source: Box::new(source),
            })?
    }

    async fn execute<T>(
        &self,
        context: &AppContext,
        tag: &str,
        operation: impl Future<Output = Result<T, CachingError>>,
    ) -> Result<T, CachingError> {
        error::finish(
            self.logger.as_ref(),
            context,
            tag,
            Self::bounded(self.operation_timeout, operation).await,
        )
    }

    pub(crate) async fn read<T: DeserializeOwned>(
        &self,
        context: &AppContext,
        lookup: CacheQuery,
        operation: (PayloadKind, &str),
    ) -> Result<CacheRead<T>, CachingError> {
        self.execute(context, operation.1, async {
            error::context(context)?;

            let info = query::describe(&lookup)?;
            if info.kind != operation.0 {
                return Err(CachingError::BadArgs);
            }

            let keys = info
                .families
                .iter()
                .map(|family| self.key_builder.revision(context, *family))
                .collect::<Result<Vec<_>, _>>()?;

            let candidates = info
                .families
                .iter()
                .map(|_| Self::token())
                .collect::<Result<Vec<_>, _>>()?;

            let mut connection = self.connection.clone();
            let script = scripts::initialize();
            let mut invocation = script.prepare_invoke();

            for key in &keys {
                invocation.key(key);
            }

            for token in candidates {
                invocation.arg(token);
            }

            let tokens: Vec<String> = invocation
                .invoke_async(&mut connection)
                .await
                .map_err(error::redis)?;
            if tokens.len() != info.families.len() {
                return Err(error::invalid_payload());
            }

            let stamp = CacheStamp {
                key: CacheEntryKey {
                    query: lookup,
                    revisions: info
                        .families
                        .into_iter()
                        .zip(tokens)
                        .map(|(family, token)| CacheRevision { family, token })
                        .collect(),
                },
            };

            let entry = self.key_builder.entry(context, &stamp.key)?;
            let script = scripts::read();
            let mut invocation = script.prepare_invoke();

            for key in &keys {
                invocation.key(key);
            }

            invocation.key(entry);
            for revision in &stamp.key.revisions {
                invocation.arg(&revision.token);
            }

            let bytes: Option<Vec<u8>> = invocation
                .invoke_async(&mut connection)
                .await
                .map_err(error::redis)?;

            let value = bytes
                .map(|bytes| {
                    let value = serde_json::from_slice(&bytes).map_err(error::failure)?;

                    validation::payload(&stamp.key.query, operation.0, &value)
                        .map_err(|_| error::invalid_payload())?;

                    serde_json::from_value(value).map_err(error::failure)
                })
                .transpose()?;

            Ok(CacheRead { value, stamp })
        })
        .await
    }

    pub(crate) async fn store<T: Serialize + ?Sized>(
        &self,
        context: &AppContext,
        stamp: &CacheStamp,
        value: &T,
        ttl: Duration,
        operation: (PayloadKind, &str),
    ) -> Result<CacheStoreOutcome, CachingError> {
        self.execute(context, operation.1, async {
            error::context(context)?;

            let info = query::describe(&stamp.key.query)?;
            if info.kind != operation.0 {
                return Err(CachingError::BadArgs);
            }

            let entry = self.key_builder.entry(context, &stamp.key)?;
            let ttl = validation::ttl(ttl)?;

            let value = serde_json::to_value(value).map_err(error::failure)?;
            validation::payload(&stamp.key.query, operation.0, &value)?;

            let bytes = serde_json::to_vec(&value).map_err(error::failure)?;

            let mut revisions: Vec<_> = stamp.key.revisions.iter().collect();
            revisions.sort_by_key(|revision| revision.family);

            // Validate dependencies even when a custom key builder is injected.
            if revisions.len() != info.families.len()
                || revisions
                    .iter()
                    .zip(&info.families)
                    .any(|(revision, family)| {
                        revision.family != *family || revision.token.is_empty()
                    })
            {
                return Err(CachingError::BadArgs);
            }

            let script = scripts::store();
            let mut invocation = script.prepare_invoke();

            for revision in &revisions {
                invocation.key(self.key_builder.revision(context, revision.family)?);
            }

            invocation.key(entry);
            for revision in revisions {
                invocation.arg(&revision.token);
            }

            invocation.arg(bytes).arg(ttl);

            let stored: i64 = invocation
                .invoke_async(&mut self.connection.clone())
                .await
                .map_err(error::redis)?;

            match stored {
                1 => Ok(CacheStoreOutcome::Stored),
                0 => Ok(CacheStoreOutcome::Superseded),
                _ => Err(error::invalid_payload()),
            }
        })
        .await
    }

    pub(crate) async fn invalidate(
        &self,
        context: &AppContext,
        families: &[CacheFamily],
    ) -> Result<(), CachingError> {
        self.execute(context, "caching/invalidation/redis/invalidate", async {
            error::context(context)?;

            let mut families = families.to_vec();
            families.sort();
            families.dedup();

            if families.is_empty() {
                return Ok(());
            }

            let script = scripts::invalidate();
            let mut invocation = script.prepare_invoke();

            for family in &families {
                invocation.key(self.key_builder.revision(context, *family)?);
            }

            for _ in families {
                invocation.arg(Self::token()?);
            }

            let changed: i64 = invocation
                .invoke_async(&mut self.connection.clone())
                .await
                .map_err(error::redis)?;
            if changed != 1 {
                return Err(error::invalid_payload());
            }

            Ok(())
        })
        .await
    }

    fn token() -> Result<String, CachingError> {
        let mut bytes = [0; 32];
        getrandom::fill(&mut bytes).map_err(error::failure)?;

        Ok(hex::encode(bytes))
    }
}
