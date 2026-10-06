use crate::domain::models::{AccessClaims, AppContext};
use crate::domain::usecases::{
    UsecaseFuture,
    shared::{RolePermissionDetails, RolePermissionWithPermission, RolePermissionWithRole},
};

pub trait RolePermission: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        input: CreateRequest,
    ) -> UsecaseFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        id: i64,
    ) -> UsecaseFuture<'a, RolePermissionDetails>;

    fn read_by_role_id_and_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        role_id: i64,
        permission_id: i64,
    ) -> UsecaseFuture<'a, RolePermissionDetails>;

    fn read_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        role_id: i64,
    ) -> UsecaseFuture<'a, Vec<RolePermissionWithPermission>>;

    fn read_by_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        permission_id: i64,
    ) -> UsecaseFuture<'a, Vec<RolePermissionWithRole>>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        id: i64,
    ) -> UsecaseFuture<'a, ()>;

    fn delete_by_role_id_or_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        role_id: Option<i64>,
        permission_id: Option<i64>,
    ) -> UsecaseFuture<'a, ()>;
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CreateRequest {
    pub role_id: i64,
    pub permission_id: i64,
}
