use serde_json::Value;

use crate::domain::models::{AccessClaims, AppContext, Space as SpaceEntity};
use crate::domain::usecases::{UsecaseFuture, shared::Page};

pub trait Space: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        input: CreateRequest,
    ) -> UsecaseFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        id: i64,
    ) -> UsecaseFuture<'a, SpaceEntity>;

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        slug: &'a str,
    ) -> UsecaseFuture<'a, SpaceEntity>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        filter: FilterRequest,
    ) -> UsecaseFuture<'a, Page<SpaceEntity>>;

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

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CreateRequest {
    pub slug: String,
    pub name: String,
    pub description: Option<String>,
    pub is_active: Option<bool>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct FilterRequest {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
    pub is_active: Option<bool>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateRequest {
    pub name: Option<String>,
    pub description: Option<String>,
    pub is_active: Option<bool>,
    pub preferences: Option<Value>,
}
