use crate::domain::models::{AppContext, Permission, Role, RolePermission as RolePermissionEntity};

use super::RepositoryFuture;

pub trait RolePermission: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        input: CreateRolePermission,
    ) -> RepositoryFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> RepositoryFuture<'a, RolePermissionDetails>;

    fn read_by_role_id_and_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        role_id: i64,
        permission_id: i64,
    ) -> RepositoryFuture<'a, RolePermissionDetails>;

    fn read_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        role_id: i64,
    ) -> RepositoryFuture<'a, Vec<RolePermissionWithPermission>>;

    fn read_by_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        permission_id: i64,
    ) -> RepositoryFuture<'a, Vec<RolePermissionWithRole>>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> RepositoryFuture<'a, ()>;

    fn delete_by_role_id_or_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        role_id: Option<i64>,
        permission_id: Option<i64>,
    ) -> RepositoryFuture<'a, ()>;
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CreateRolePermission {
    pub role_id: i64,
    pub permission_id: i64,
    pub by: Option<i64>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct RolePermissionDetails {
    pub role_permission: RolePermissionEntity,
    pub role: Role,
    pub permission: Permission,
}

#[derive(Debug, Clone, PartialEq)]
pub struct RolePermissionWithPermission {
    pub role_permission: RolePermissionEntity,
    pub permission: Permission,
}

#[derive(Debug, Clone, PartialEq)]
pub struct RolePermissionWithRole {
    pub role_permission: RolePermissionEntity,
    pub role: Role,
}
