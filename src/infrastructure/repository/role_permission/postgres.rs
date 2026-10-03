use std::sync::Arc;

use futures_util::TryStreamExt;

use mate_pgdt::{
    Pgdt,
    sqlx::{self, Row, postgres::PgRow},
};

use crate::domain::{
    contracts::{
        repository::{
            CreateRolePermission, RepositoryFuture, RolePermission, RolePermissionDetails,
            RolePermissionWithPermission, RolePermissionWithRole,
        },
        utility::Logger,
    },
    models::{AppContext, RepositoryError},
};

use super::{
    super::shared::{
        error::{OperationError, finish, require_affected},
        scan,
    },
    postgres_query as query,
};

pub struct PostgresRolePermission {
    database: Pgdt,
    logger: Arc<dyn Logger>,
}

impl PostgresRolePermission {
    pub fn new(database: Pgdt, logger: Arc<dyn Logger>) -> Self {
        Self { database, logger }
    }
}

impl RolePermission for PostgresRolePermission {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        input: CreateRolePermission,
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
                "repository/role_permission/postgres/create",
                result,
                || RepositoryError::RolePermissionNotFound,
            )
        })
    }

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> RepositoryFuture<'a, RolePermissionDetails> {
        Box::pin(async move {
            let result: Result<RolePermissionDetails, OperationError> = async {
                let mut query = query::read_by_id(id);
                let row = self.database.query_row(context, query.build()).await?;
                Ok(decode(&row)?)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role_permission/postgres/read_by_id",
                result,
                || RepositoryError::RolePermissionNotFound,
            )
        })
    }

    fn read_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        role_id: i64,
    ) -> RepositoryFuture<'a, Vec<RolePermissionWithPermission>> {
        Box::pin(async move {
            let result: Result<Vec<RolePermissionWithPermission>, OperationError> = async {
                let mut query = query::read_by_role_id(role_id);
                let mut rows = self.database.query(context, query.build());
                let mut items = Vec::new();
                while let Some(row) = rows.try_next().await? {
                    items.push(RolePermissionWithPermission {
                        role_permission: scan::role_permission::decode(&row, "rp_")?,
                        permission: scan::permission::decode(&row, "p_")?,
                    });
                }
                Ok(items)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role_permission/postgres/read_by_role_id",
                result,
                || RepositoryError::RolePermissionNotFound,
            )
        })
    }

    fn read_by_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        permission_id: i64,
    ) -> RepositoryFuture<'a, Vec<RolePermissionWithRole>> {
        Box::pin(async move {
            let result: Result<Vec<RolePermissionWithRole>, OperationError> = async {
                let mut query = query::read_by_permission_id(permission_id);
                let mut rows = self.database.query(context, query.build());
                let mut items = Vec::new();
                while let Some(row) = rows.try_next().await? {
                    items.push(RolePermissionWithRole {
                        role_permission: scan::role_permission::decode(&row, "rp_")?,
                        role: scan::role::decode(&row, "r_")?,
                    });
                }
                Ok(items)
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role_permission/postgres/read_by_permission_id",
                result,
                || RepositoryError::RolePermissionNotFound,
            )
        })
    }

    fn delete_by_id<'a>(&'a self, context: &'a AppContext, id: i64) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query = query::delete_by_id(id);
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::RolePermissionNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role_permission/postgres/delete_by_id",
                result,
                || RepositoryError::RolePermissionNotFound,
            )
        })
    }

    fn delete_by_role_id_or_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        role_id: Option<i64>,
        permission_id: Option<i64>,
    ) -> RepositoryFuture<'a, ()> {
        Box::pin(async move {
            let result: Result<(), OperationError> = async {
                let mut query = query::delete_by_role_id_or_permission_id(role_id, permission_id)?;
                require_affected(
                    self.database
                        .exec(context, query.build())
                        .await?
                        .rows_affected(),
                    || RepositoryError::RolePermissionNotFound,
                )
            }
            .await;
            finish(
                self.logger.as_ref(),
                context,
                "repository/role_permission/postgres/delete_by_role_id_or_permission_id",
                result,
                || RepositoryError::RolePermissionNotFound,
            )
        })
    }
}

fn decode(row: &PgRow) -> Result<RolePermissionDetails, sqlx::Error> {
    Ok(RolePermissionDetails {
        role_permission: scan::role_permission::decode(row, "rp_")?,
        role: scan::role::decode(row, "r_")?,
        permission: scan::permission::decode(row, "p_")?,
    })
}
