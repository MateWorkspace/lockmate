use std::sync::Arc;

use futures_util::TryStreamExt;

use mate_pgdt::{Pgdt, sqlx::Row};

use crate::domain::{
    contracts::{
        repository::{
            CreateSpaceMember, RepositoryFuture, SpaceMember, SpaceMemberFilter,
            SpaceMemberWithUser, UpdateSpaceMember,
        },
        utility::Logger,
    },
    models::{AppContext, RepositoryError, SpaceMember as SpaceMemberEntity},
};

use super::{
    super::shared::{
        error::{OperationError, finish, require_affected},
        scan,
    },
    postgres_query as query,
};

pub struct PostgresSpaceMember {
    database: Pgdt,
    logger: Arc<dyn Logger>,
}

impl PostgresSpaceMember {
    pub fn new(database: Pgdt, logger: Arc<dyn Logger>) -> Self {
        Self { database, logger }
    }
}

impl SpaceMember for PostgresSpaceMember {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        input: CreateSpaceMember,
    ) -> RepositoryFuture<'a, i64> {
        Box::pin(async move {
            let result: Result<i64, OperationError> = async {
                self.database
                    .with_tx(context, |context| async move {
                        let mut parents = query::lock_parents(space_id, input.user_id);
                        self.database
                            .query_optional(&context, parents.build())
                            .await?
                            .ok_or(RepositoryError::BadArgs)?;
                        let mut lock = query::lock_default(space_id);
                        self.database.exec(&context, lock.build()).await?;
                        let mut default = query::read_default_for_share(space_id);
                        let role_id: i64 = self
                            .database
                            .query_optional(&context, default.build())
                            .await?
                            .ok_or(RepositoryError::RoleNotFound)?
                            .try_get("id")?;
                        let mut create = query::create(space_id, &input);
                        let member_id: i64 = self
                            .database
                            .query_optional(&context, create.build())
                            .await?
                            .ok_or(RepositoryError::BadArgs)?
                            .try_get("id")?;
                        let mut assign =
                            query::assign_default(space_id, member_id, role_id, input.by);
                        self.database.exec(&context, assign.build()).await?;
                        Ok::<_, OperationError>(member_id)
                    })
                    .await
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space_member/postgres/create",
                result,
                || RepositoryError::SpaceMemberNotFound,
            )
        })
    }

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> RepositoryFuture<'a, SpaceMemberEntity> {
        Box::pin(async move {
            let result: Result<SpaceMemberEntity, OperationError> = async {
                let mut query = query::read_by_id(space_id, id);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::space_member::decode(&row, "m_")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space_member/postgres/read_by_id",
                result,
                || RepositoryError::SpaceMemberNotFound,
            )
        })
    }

    fn read_by_user_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        user_id: i64,
    ) -> RepositoryFuture<'a, SpaceMemberEntity> {
        Box::pin(async move {
            let result: Result<SpaceMemberEntity, OperationError> = async {
                let mut query = query::read_by_user_id(space_id, user_id);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(scan::space_member::decode(&row, "m_")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space_member/postgres/read_by_user_id",
                result,
                || RepositoryError::SpaceMemberNotFound,
            )
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: SpaceMemberFilter,
    ) -> RepositoryFuture<'a, (Vec<SpaceMemberWithUser>, i64)> {
        Box::pin(async move {
            let result: Result<(Vec<SpaceMemberWithUser>, i64), OperationError> = async {
                let (mut count, mut query) = query::read_by_filter(space_id, &filter)?;
                let total: i64 = self
                    .database
                    .query_row(context, count.build())
                    .await?
                    .try_get("total")?;
                let mut items = Vec::new();
                if total > 0 {
                    let mut rows = self.database.query(context, query.build());
                    while let Some(row) = rows.try_next().await? {
                        items.push(SpaceMemberWithUser {
                            member: scan::space_member::decode(&row, "m_")?,
                            user: scan::user::decode(&row, "u_")?,
                        });
                    }
                }
                Ok((items, total))
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space_member/postgres/read_by_filter",
                result,
                || RepositoryError::SpaceMemberNotFound,
            )
        })
    }

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
        input: UpdateSpaceMember,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query = query::update_by_id(space_id, id, &input);
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::SpaceMemberNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space_member/postgres/update_by_id",
                result,
                || RepositoryError::SpaceMemberNotFound,
            )
        })
    }

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
        by: Option<i64>,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query = query::delete_by_id(space_id, id, by);
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::SpaceMemberNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/space_member/postgres/delete_by_id",
                result,
                || RepositoryError::SpaceMemberNotFound,
            )
        })
    }
}
