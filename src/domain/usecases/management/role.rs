use serde_json::Value;

use crate::domain::models::{AccessClaims, AppContext, Role as RoleEntity};
use crate::domain::usecases::{UsecaseFuture, shared::Page};

pub trait Role: Send + Sync {
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
    ) -> UsecaseFuture<'a, RoleEntity>;

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        slug: &'a str,
    ) -> UsecaseFuture<'a, RoleEntity>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        filter: FilterRequest,
    ) -> UsecaseFuture<'a, Page<RoleEntity>>;

    fn read_default<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
    ) -> UsecaseFuture<'a, RoleEntity>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        id: i64,
        input: UpdateRequest,
    ) -> UsecaseFuture<'a, ()>;

    fn set_default<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        id: i64,
    ) -> UsecaseFuture<'a, ()>;

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
    pub slug: String,
    pub name: String,
    pub description: Option<String>,
    pub is_default: Option<bool>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct FilterRequest {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
    pub is_default: Option<bool>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateRequest {
    pub name: Option<String>,
    pub description: Option<String>,
    pub is_default: Option<bool>,
    pub preferences: Option<Value>,
}
