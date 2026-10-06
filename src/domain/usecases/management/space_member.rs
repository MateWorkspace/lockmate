use serde_json::Value;

use crate::domain::models::{
    AccessClaims, AppContext, Permission, Role, SpaceMember as SpaceMemberEntity,
};
use crate::domain::usecases::{
    UsecaseFuture,
    shared::{Page, SpaceMemberWithUser},
};

pub trait SpaceMember: Send + Sync {
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
    ) -> UsecaseFuture<'a, SpaceMemberEntity>;

    fn read_by_user_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        user_id: i64,
    ) -> UsecaseFuture<'a, SpaceMemberEntity>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        filter: FilterRequest,
    ) -> UsecaseFuture<'a, Page<SpaceMemberWithUser>>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        id: i64,
        input: UpdateRequest,
    ) -> UsecaseFuture<'a, ()>;

    fn read_effective_roles_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        member_id: i64,
    ) -> UsecaseFuture<'a, Vec<Role>>;

    fn read_effective_permissions_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        member_id: i64,
    ) -> UsecaseFuture<'a, Vec<Permission>>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        id: i64,
    ) -> UsecaseFuture<'a, ()>;
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CreateRequest {
    pub user_id: i64,
    pub is_active: Option<bool>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct FilterRequest {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
    pub user_id: Option<i64>,
    pub role_id: Option<i64>,
    pub is_active: Option<bool>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateRequest {
    pub is_active: Option<bool>,
    pub preferences: Option<Value>,
}
