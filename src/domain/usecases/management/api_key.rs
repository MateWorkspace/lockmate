use serde_json::Value;

use crate::domain::models::{AccessClaims, AppContext};
use crate::domain::usecases::{
    UsecaseFuture,
    shared::{ApiKeyView, CreatedApiKey, Page},
};

pub trait ApiKey: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        input: CreateRequest,
    ) -> UsecaseFuture<'a, CreatedApiKey>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        id: i64,
    ) -> UsecaseFuture<'a, ApiKeyView>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        filter: FilterRequest,
    ) -> UsecaseFuture<'a, Page<ApiKeyView>>;

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

#[derive(Clone, PartialEq, Eq)]
pub struct CreateRequest {
    pub member_id: i64,
    pub name: String,
    pub description: Option<String>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct FilterRequest {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
    pub member_id: Option<i64>,
    pub user_id: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateRequest {
    pub name: Option<String>,
    pub description: Option<String>,
    pub preferences: Option<Value>,
}
