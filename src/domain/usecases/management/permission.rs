use serde_json::Value;

use crate::domain::models::{AccessClaims, AppContext, Permission as PermissionEntity};
use crate::domain::usecases::{UsecaseFuture, shared::Page};

pub trait Permission: Send + Sync {
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
    ) -> UsecaseFuture<'a, PermissionEntity>;

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        slug: &'a str,
    ) -> UsecaseFuture<'a, PermissionEntity>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        filter: FilterRequest,
    ) -> UsecaseFuture<'a, Page<PermissionEntity>>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        id: i64,
        input: UpdateRequest,
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
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct FilterRequest {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateRequest {
    pub name: Option<String>,
    pub description: Option<String>,
    pub preferences: Option<Value>,
}
