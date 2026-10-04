use std::time::Duration;

use crate::domain::contracts::repository::ApiKeyFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, CachedApiKey,
};

use super::super::{RedisBackend, shared::query::PayloadKind};
use crate::domain::contracts::caching::{ApiKey, CachingFuture};
use crate::domain::models::CacheQuery;

pub struct RedisApiKey {
    rdb: RedisBackend,
}

impl RedisApiKey {
    pub fn new(rdb: RedisBackend) -> Self {
        Self { rdb }
    }
}

impl ApiKey for RedisApiKey {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedApiKey>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::ApiKeyById { space_id, id },
                    (
                        PayloadKind::ApiKeyRecord,
                        "caching/api_key/redis/read_by_id",
                    ),
                )
                .await
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: ApiKeyFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<CachedApiKey>>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::ApiKeyByFilter { space_id, filter },
                    (
                        PayloadKind::ApiKeyPage,
                        "caching/api_key/redis/read_by_filter",
                    ),
                )
                .await
        })
    }

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachedApiKey,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome> {
        Box::pin(async move {
            self.rdb
                .store(
                    context,
                    stamp,
                    value,
                    ttl,
                    (
                        PayloadKind::ApiKeyRecord,
                        "caching/api_key/redis/store_record",
                    ),
                )
                .await
        })
    }

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<CachedApiKey>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome> {
        Box::pin(async move {
            self.rdb
                .store(
                    context,
                    stamp,
                    value,
                    ttl,
                    (PayloadKind::ApiKeyPage, "caching/api_key/redis/store_page"),
                )
                .await
        })
    }
}
