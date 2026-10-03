use std::sync::Arc;

use futures_util::TryStreamExt;

use mate_pgdt::{Pgdt, sqlx::Row};

use crate::domain::{
    contracts::{
        repository::{ApiKey, ApiKeyFilter, CreateApiKey, RepositoryFuture, UpdateApiKey},
        utility::Logger,
    },
    models::{ApiKey as ApiKeyEntity, AppContext, RepositoryError},
};

use super::{
    super::shared::{
        error::{OperationError, finish, require_affected},
        scan,
    },
    postgres_query as query,
};

pub struct PostgresApiKey {
    database: Pgdt,
    logger: Arc<dyn Logger>,
}

impl PostgresApiKey {
    pub fn new(database: Pgdt, logger: Arc<dyn Logger>) -> Self {
        Self { database, logger }
    }
}

impl ApiKey for PostgresApiKey {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        input: CreateApiKey,
    ) -> RepositoryFuture<'a, i64> {
        Box::pin(async move {
            let result: Result<i64, OperationError> = async {
                let mut query = query::create(&input);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(row.try_get("id")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/api_key/postgres/create",
                result,
                || RepositoryError::UserApiKeyNotFound,
            )
        })
    }

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> RepositoryFuture<'a, ApiKeyEntity> {
        Box::pin(async move {
            let result: Result<ApiKeyEntity, OperationError> = async {
                let mut query = query::read_by_id(id);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::api_key::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/api_key/postgres/read_by_id",
                result,
                || RepositoryError::UserApiKeyNotFound,
            )
        })
    }

    fn read_by_hash<'a>(
        &'a self,
        context: &'a AppContext,
        hash: &'a str,
    ) -> RepositoryFuture<'a, ApiKeyEntity> {
        Box::pin(async move {
            let result: Result<ApiKeyEntity, OperationError> = async {
                let mut query = query::read_by_hash(hash);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::api_key::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/api_key/postgres/read_by_hash",
                result,
                || RepositoryError::UserApiKeyNotFound,
            )
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: ApiKeyFilter,
    ) -> RepositoryFuture<'a, (Vec<ApiKeyEntity>, i64)> {
        Box::pin(async move {
            let result: Result<(Vec<ApiKeyEntity>, i64), OperationError> = async {
                let (mut count, mut query) = query::read_by_filter(&filter)?;
                let total: i64 = self
                    .database
                    .query_row(context, count.build())
                    .await?
                    .try_get("total")?;
                let mut items = Vec::new();
                if total > 0 {
                    let mut rows = self.database.query(context, query.build());
                    while let Some(row) = rows.try_next().await? {
                        items.push(scan::api_key::decode(&row, "")?);
                    }
                }
                Ok((items, total))
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/api_key/postgres/read_by_filter",
                result,
                || RepositoryError::UserApiKeyNotFound,
            )
        })
    }

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        input: UpdateApiKey,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query = query::update_by_id(id, &input);
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::UserApiKeyNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/api_key/postgres/update_by_id",
                result,
                || RepositoryError::UserApiKeyNotFound,
            )
        })
    }

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        by: Option<i64>,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query = query::delete_by_id(id, by);
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::UserApiKeyNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/api_key/postgres/delete_by_id",
                result,
                || RepositoryError::UserApiKeyNotFound,
            )
        })
    }
}
