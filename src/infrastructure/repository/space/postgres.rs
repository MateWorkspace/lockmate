use std::sync::Arc;

use futures_util::TryStreamExt;

use mate_pgdt::{Pgdt, sqlx::Row};

use crate::domain::{
    contracts::{
        repository::{CreateSpace, RepositoryFuture, Space, SpaceFilter, UpdateSpace},
        utility::Logger,
    },
    models::{AppContext, RepositoryError, Space as SpaceEntity},
};

use super::{
    super::shared::{
        error::{OperationError, finish, require_affected},
        scan,
    },
    postgres_query as query,
};

pub struct PostgresSpace {
    database: Pgdt,
    logger: Arc<dyn Logger>,
}

impl PostgresSpace {
    pub fn new(database: Pgdt, logger: Arc<dyn Logger>) -> Self {
        Self { database, logger }
    }
}

impl Space for PostgresSpace {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        input: CreateSpace,
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
                "repository/space/postgres/create",
                result,
                || RepositoryError::SpaceNotFound,
            )
        })
    }

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> RepositoryFuture<'a, SpaceEntity> {
        Box::pin(async move {
            let result: Result<SpaceEntity, OperationError> = async {
                let mut query = query::read_by_id(id);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::space::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space/postgres/read_by_id",
                result,
                || RepositoryError::SpaceNotFound,
            )
        })
    }

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        slug: &'a str,
    ) -> RepositoryFuture<'a, SpaceEntity> {
        Box::pin(async move {
            let result: Result<SpaceEntity, OperationError> = async {
                let mut query = query::read_by_slug(slug);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::space::decode(&row, "")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space/postgres/read_by_slug",
                result,
                || RepositoryError::SpaceNotFound,
            )
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: SpaceFilter,
    ) -> RepositoryFuture<'a, (Vec<SpaceEntity>, i64)> {
        Box::pin(async move {
            let result: Result<(Vec<SpaceEntity>, i64), OperationError> = async {
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
                        items.push(scan::space::decode(&row, "")?);
                    }
                }
                Ok((items, total))
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space/postgres/read_by_filter",
                result,
                || RepositoryError::SpaceNotFound,
            )
        })
    }

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        input: UpdateSpace,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query = query::update_by_id(id, &input);
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::SpaceNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space/postgres/update_by_id",
                result,
                || RepositoryError::SpaceNotFound,
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
                    || RepositoryError::SpaceNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space/postgres/delete_by_id",
                result,
                || RepositoryError::SpaceNotFound,
            )
        })
    }
}
