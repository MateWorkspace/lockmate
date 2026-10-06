use serde_json::Value;

use crate::domain::models::{AccessClaims, AppContext};
use crate::domain::usecases::{
    UsecaseFuture,
    shared::{ApiKeyView, CreatedApiKey, Page},
};

/// Read and mutate only keys belonging to the caller's current membership.
pub trait ApiKey: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        input: CreateRequest,
    ) -> UsecaseFuture<'a, CreatedApiKey>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        id: i64,
    ) -> UsecaseFuture<'a, ApiKeyView>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        filter: FilterRequest,
    ) -> UsecaseFuture<'a, Page<ApiKeyView>>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        id: i64,
        input: UpdateRequest,
    ) -> UsecaseFuture<'a, ()>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        id: i64,
    ) -> UsecaseFuture<'a, ()>;
}

#[derive(Clone, PartialEq, Eq)]
pub struct CreateRequest {
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
