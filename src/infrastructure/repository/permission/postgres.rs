use std::sync::Arc;

use futures_util::TryStreamExt;

use mate_pgdt::{Pgdt, sqlx::Row};

use crate::domain::{
    contracts::{
        repository::{
            CreatePermission, Permission, PermissionFilter, RepositoryFuture, UpdatePermission,
        },
        utility::Logger,
    },
    models::{AppContext, Permission as PermissionEntity, RepositoryError},
};

use super::{
    super::shared::{
        error::{OperationError, finish, require_affected},
        scan,
    },
    postgres_query as query,
};

pub struct PostgresPermission {
    database: Pgdt,
    logger: Arc<dyn Logger>,
}

impl PostgresPermission {
    pub fn new(database: Pgdt, logger: Arc<dyn Logger>) -> Self {
        Self { database, logger }
    }
}

impl Permission for PostgresPermission {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        input: CreatePermission,
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
                "repository/permission/postgres/create",
                result,
                || RepositoryError::PermissionNotFound,
            )
        })
    }

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> RepositoryFuture<'a, PermissionEntity> {
        Box::pin(async move {
            let result: Result<PermissionEntity, OperationError> = async {
                let mut query = query::read_by_id(id);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::permission::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/permission/postgres/read_by_id",
                result,
                || RepositoryError::PermissionNotFound,
            )
        })
    }

    fn read_by_name<'a>(
        &'a self,
        context: &'a AppContext,
        name: &'a str,
    ) -> RepositoryFuture<'a, PermissionEntity> {
        Box::pin(async move {
            let result: Result<PermissionEntity, OperationError> = async {
                let mut query = query::read_by_name(name);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::permission::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/permission/postgres/read_by_name",
                result,
                || RepositoryError::PermissionNotFound,
            )
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: PermissionFilter,
    ) -> RepositoryFuture<'a, (Vec<PermissionEntity>, i64)> {
        Box::pin(async move {
            let result: Result<(Vec<PermissionEntity>, i64), OperationError> = async {
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
                        items.push(scan::permission::decode(&row, "")?);
                    }
                }
                Ok((items, total))
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/permission/postgres/read_by_filter",
                result,
                || RepositoryError::PermissionNotFound,
            )
        })
    }

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        input: UpdatePermission,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query = query::update_by_id(id, &input);
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::PermissionNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/permission/postgres/update_by_id",
                result,
                || RepositoryError::PermissionNotFound,
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
                    || RepositoryError::PermissionNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/permission/postgres/delete_by_id",
                result,
                || RepositoryError::PermissionNotFound,
            )
        })
    }
}
