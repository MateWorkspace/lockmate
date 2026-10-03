use std::sync::Arc;

use futures_util::TryStreamExt;

use mate_pgdt::{Pgdt, sqlx::Row};

use crate::domain::{
    contracts::{
        repository::{CreateRole, RepositoryFuture, Role, RoleFilter, UpdateRole},
        utility::Logger,
    },
    models::{AppContext, RepositoryError, Role as RoleEntity},
};

use super::{
    super::shared::{
        error::{OperationError, finish, require_affected},
        scan,
    },
    postgres_query as query,
};

pub struct PostgresRole {
    database: Pgdt,
    logger: Arc<dyn Logger>,
}

impl PostgresRole {
    pub fn new(database: Pgdt, logger: Arc<dyn Logger>) -> Self {
        Self { database, logger }
    }
}

impl Role for PostgresRole {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        input: CreateRole,
    ) -> RepositoryFuture<'a, i64> {
        Box::pin(async move {
            let result: Result<i64, OperationError> = async {
                if input.is_default == Some(true) {
                    self.database
                        .with_tx(context, |context| async move {
                            let mut lock = query::lock_default();
                            self.database.exec(&context, lock.build()).await?;
                            let mut unset = query::unset_default(None, input.by);
                            self.database.exec(&context, unset.build()).await?;
                            let mut query = query::create(&input);
                            let row = self.database.query_row(&context, query.build()).await?;
                            Ok::<_, OperationError>(row.try_get("id")?)
                        })
                        .await
                } else {
                    let mut query = query::create(&input);
                    let row = self.database.query_row(context, query.build()).await?;
                    Ok(row.try_get("id")?)
                }
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role/postgres/create",
                result,
                || RepositoryError::RoleNotFound,
            )
        })
    }

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> RepositoryFuture<'a, RoleEntity> {
        Box::pin(async move {
            let result: Result<RoleEntity, OperationError> = async {
                let mut query = query::read_by_id(id);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::role::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role/postgres/read_by_id",
                result,
                || RepositoryError::RoleNotFound,
            )
        })
    }

    fn read_by_name<'a>(
        &'a self,
        context: &'a AppContext,
        name: &'a str,
    ) -> RepositoryFuture<'a, RoleEntity> {
        Box::pin(async move {
            let result: Result<RoleEntity, OperationError> = async {
                let mut query = query::read_by_name(name);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::role::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role/postgres/read_by_name",
                result,
                || RepositoryError::RoleNotFound,
            )
        })
    }

    fn read_default<'a>(&'a self, context: &'a AppContext) -> RepositoryFuture<'a, RoleEntity> {
        Box::pin(async move {
            let result: Result<RoleEntity, OperationError> = async {
                let mut query = query::read_default();
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::role::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role/postgres/read_default",
                result,
                || RepositoryError::RoleNotFound,
            )
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: RoleFilter,
    ) -> RepositoryFuture<'a, (Vec<RoleEntity>, i64)> {
        Box::pin(async move {
            let result: Result<(Vec<RoleEntity>, i64), OperationError> = async {
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
                        items.push(scan::role::decode(&row, "")?);
                    }
                }
                Ok((items, total))
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role/postgres/read_by_filter",
                result,
                || RepositoryError::RoleNotFound,
            )
        })
    }

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        input: UpdateRole,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                if input.is_default == Some(true) {
                    self.database
                        .with_tx(context, |context| async move {
                            let mut lock = query::lock_default();
                            self.database.exec(&context, lock.build()).await?;
                            let mut unset = query::unset_default(Some(id), input.by);
                            self.database.exec(&context, unset.build()).await?;
                            let mut query = query::update_by_id(id, &input);
                            require_affected(
                                self.database
                                    .exec(&context, query.build())
                                    .await?
                                    .rows_affected(),
                                || RepositoryError::RoleNotFound,
                            )
                        })
                        .await
                } else {
                    let mut query = query::update_by_id(id, &input);
                    require_affected(
                        self.database
                            .exec(context, query.build())
                            .await?
                            .rows_affected(),
                        || RepositoryError::RoleNotFound,
                    )
                }
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role/postgres/update_by_id",
                result,
                || RepositoryError::RoleNotFound,
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
                    || RepositoryError::RoleNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role/postgres/delete_by_id",
                result,
                || RepositoryError::RoleNotFound,
            )
        })
    }
}
