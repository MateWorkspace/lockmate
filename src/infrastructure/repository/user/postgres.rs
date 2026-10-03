use std::sync::Arc;

use futures_util::TryStreamExt;

use mate_pgdt::{Pgdt, sqlx::Row};

use crate::domain::{
    contracts::{
        repository::{CreateUser, RepositoryFuture, UpdateUser, User, UserFilter},
        utility::Logger,
    },
    models::{AppContext, RepositoryError, User as UserEntity},
};

use super::{
    super::shared::{
        error::{OperationError, finish, require_affected},
        scan,
    },
    postgres_query as query,
};

pub struct PostgresUser {
    database: Pgdt,
    logger: Arc<dyn Logger>,
}

impl PostgresUser {
    pub fn new(database: Pgdt, logger: Arc<dyn Logger>) -> Self {
        Self { database, logger }
    }
}

impl User for PostgresUser {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        input: CreateUser,
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
                "repository/user/postgres/create",
                result,
                || RepositoryError::UserNotFound,
            )
        })
    }

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> RepositoryFuture<'a, UserEntity> {
        Box::pin(async move {
            let result: Result<UserEntity, OperationError> = async {
                let mut query = query::read_by_id(id);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::user::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/user/postgres/read_by_id",
                result,
                || RepositoryError::UserNotFound,
            )
        })
    }

    fn read_by_username<'a>(
        &'a self,
        context: &'a AppContext,
        username: &'a str,
    ) -> RepositoryFuture<'a, UserEntity> {
        Box::pin(async move {
            let result: Result<UserEntity, OperationError> = async {
                let mut query = query::read_by_username(username);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::user::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/user/postgres/read_by_username",
                result,
                || RepositoryError::UserNotFound,
            )
        })
    }

    fn read_by_email<'a>(
        &'a self,
        context: &'a AppContext,
        email: &'a str,
    ) -> RepositoryFuture<'a, UserEntity> {
        Box::pin(async move {
            let result: Result<UserEntity, OperationError> = async {
                let mut query = query::read_by_email(email);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::user::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/user/postgres/read_by_email",
                result,
                || RepositoryError::UserNotFound,
            )
        })
    }

    fn read_by_phone<'a>(
        &'a self,
        context: &'a AppContext,
        phone: &'a str,
    ) -> RepositoryFuture<'a, UserEntity> {
        Box::pin(async move {
            let result: Result<UserEntity, OperationError> = async {
                let mut query = query::read_by_phone(phone);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::user::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/user/postgres/read_by_phone",
                result,
                || RepositoryError::UserNotFound,
            )
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: UserFilter,
    ) -> RepositoryFuture<'a, (Vec<UserEntity>, i64)> {
        Box::pin(async move {
            let result: Result<(Vec<UserEntity>, i64), OperationError> = async {
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
                        items.push(scan::user::decode(&row, "")?);
                    }
                }
                Ok((items, total))
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/user/postgres/read_by_filter",
                result,
                || RepositoryError::UserNotFound,
            )
        })
    }

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        input: UpdateUser,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query = query::update_by_id(id, &input);
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::UserNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/user/postgres/update_by_id",
                result,
                || RepositoryError::UserNotFound,
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
                    || RepositoryError::UserNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/user/postgres/delete_by_id",
                result,
                || RepositoryError::UserNotFound,
            )
        })
    }
}
