use std::sync::Arc;

use futures_util::TryStreamExt;

use mate_pgdt::{Pgdt, sqlx::Row};

use crate::domain::{
    contracts::{
        repository::{
            CreateMemberRole, MemberRole, MemberRoleDetails, MemberRoleWithMember,
            MemberRoleWithRole, RepositoryFuture,
        },
        utility::Logger,
    },
    models::{AppContext, Permission, RepositoryError, Role},
};

use super::{
    super::shared::{
        error::{OperationError, finish, require_affected},
        scan,
    },
    postgres_query as query,
};

pub struct PostgresMemberRole {
    database: Pgdt,
    logger: Arc<dyn Logger>,
}

impl PostgresMemberRole {
    pub fn new(database: Pgdt, logger: Arc<dyn Logger>) -> Self {
        Self { database, logger }
    }
}

impl MemberRole for PostgresMemberRole {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        input: CreateMemberRole,
    ) -> RepositoryFuture<'a, i64> {
        Box::pin(async move {
            let result: Result<i64, OperationError> = async {
                let mut query = query::create(space_id, &input);
                let row = self
                    .database
                    .query_optional(context, query.build())
                    .await?
                    .ok_or(RepositoryError::BadArgs)?;
                Ok(row.try_get("id")?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/member_role/postgres/create",
                result,
                || RepositoryError::MemberRoleNotFound,
            )
        })
    }

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> RepositoryFuture<'a, MemberRoleDetails> {
        Box::pin(async move {
            let result: Result<MemberRoleDetails, OperationError> = async {
                let mut query = query::read_by_id(space_id, id);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(MemberRoleDetails {
                    member_role: scan::member_role::decode(&row, "mr_")?,
                    member: scan::space_member::decode(&row, "m_")?,
                    role: scan::role::decode(&row, "r_")?,
                })
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/member_role/postgres/read_by_id",
                result,
                || RepositoryError::MemberRoleNotFound,
            )
        })
    }

    fn read_by_member_id_and_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
        role_id: i64,
    ) -> RepositoryFuture<'a, MemberRoleDetails> {
        Box::pin(async move {
            let result: Result<MemberRoleDetails, OperationError> = async {
                let mut query = query::read_by_member_id_and_role_id(space_id, member_id, role_id);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(MemberRoleDetails {
                    member_role: scan::member_role::decode(&row, "mr_")?,
                    member: scan::space_member::decode(&row, "m_")?,
                    role: scan::role::decode(&row, "r_")?,
                })
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/member_role/postgres/read_by_member_id_and_role_id",
                result,
                || RepositoryError::MemberRoleNotFound,
            )
        })
    }

    fn read_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        role_id: i64,
    ) -> RepositoryFuture<'a, Vec<MemberRoleWithMember>> {
        Box::pin(async move {
            let result: Result<Vec<MemberRoleWithMember>, OperationError> = async {
                let mut query = query::read_by_role_id(space_id, role_id);
                let mut rows = self.database.query(context, query.build());
                let mut items = Vec::new();
                while let Some(row) = rows.try_next().await? {
                    items.push(MemberRoleWithMember {
                        member_role: scan::member_role::decode(&row, "mr_")?,
                        member: scan::space_member::decode(&row, "m_")?,
                    });
                }
                Ok(items)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/member_role/postgres/read_by_role_id",
                result,
                || RepositoryError::MemberRoleNotFound,
            )
        })
    }

    fn read_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
    ) -> RepositoryFuture<'a, Vec<MemberRoleWithRole>> {
        Box::pin(async move {
            let result: Result<Vec<MemberRoleWithRole>, OperationError> = async {
                let mut query = query::read_by_member_id(space_id, member_id);
                let mut rows = self.database.query(context, query.build());
                let mut items = Vec::new();
                while let Some(row) = rows.try_next().await? {
                    items.push(MemberRoleWithRole {
                        member_role: scan::member_role::decode(&row, "mr_")?,
                        role: scan::role::decode(&row, "r_")?,
                    });
                }
                Ok(items)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/member_role/postgres/read_by_member_id",
                result,
                || RepositoryError::MemberRoleNotFound,
            )
        })
    }

    fn read_effective_roles_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
    ) -> RepositoryFuture<'a, Vec<Role>> {
        Box::pin(async move {
            let result: Result<Vec<Role>, OperationError> = async {
                let mut state = query::access_state(space_id, member_id);
                let state = self
                    .database
                    .query_optional(context, state.build())
                    .await?
                    .ok_or(RepositoryError::SpaceMemberNotFound)?;
                if !state.try_get::<bool, _>("space_active")?
                    || !state.try_get::<bool, _>("member_active")?
                {
                    return Err(RepositoryError::BadState.into());
                }
                let mut query = query::read_effective_roles_by_member_id(space_id, member_id);
                let mut rows = self.database.query(context, query.build());
                let mut items = Vec::new();
                while let Some(row) = rows.try_next().await? {
                    items.push(scan::role::decode(&row, "")?);
                }
                Ok(items)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/member_role/postgres/read_effective_roles_by_member_id",
                result,
                || RepositoryError::MemberRoleNotFound,
            )
        })
    }

    fn read_effective_permissions_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
    ) -> RepositoryFuture<'a, Vec<Permission>> {
        Box::pin(async move {
            let result: Result<Vec<Permission>, OperationError> = async {
                let mut state = query::access_state(space_id, member_id);
                let state = self
                    .database
                    .query_optional(context, state.build())
                    .await?
                    .ok_or(RepositoryError::SpaceMemberNotFound)?;
                if !state.try_get::<bool, _>("space_active")?
                    || !state.try_get::<bool, _>("member_active")?
                {
                    return Err(RepositoryError::BadState.into());
                }
                let mut query = query::read_effective_permissions_by_member_id(space_id, member_id);
                let mut rows = self.database.query(context, query.build());
                let mut items = Vec::new();
                while let Some(row) = rows.try_next().await? {
                    items.push(scan::permission::decode(&row, "")?);
                }
                Ok(items)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/member_role/postgres/read_effective_permissions_by_member_id",
                result,
                || RepositoryError::MemberRoleNotFound,
            )
        })
    }

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query = query::delete_by_id(space_id, id);
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::MemberRoleNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/member_role/postgres/delete_by_id",
                result,
                || RepositoryError::MemberRoleNotFound,
            )
        })
    }

    fn delete_by_member_id_or_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: Option<i64>,
        role_id: Option<i64>,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query =
                    query::delete_by_member_id_or_role_id(space_id, member_id, role_id)?;
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::MemberRoleNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/member_role/postgres/delete_by_member_id_or_role_id",
                result,
                || RepositoryError::MemberRoleNotFound,
            )
        })
    }
}
