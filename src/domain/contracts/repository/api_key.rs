use serde_json::Value;

use crate::domain::models::{ApiKey as ApiKeyEntity, AppContext};

use super::RepositoryFuture;

pub trait ApiKey: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        input: CreateApiKey,
    ) -> RepositoryFuture<'a, i64>;
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> RepositoryFuture<'a, ApiKeyEntity>;
    fn read_by_hash<'a>(
        &'a self,
        context: &'a AppContext,
        hash: &'a str,
    ) -> RepositoryFuture<'a, ApiKeyEntity>;
    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: ApiKeyFilter,
    ) -> RepositoryFuture<'a, (Vec<ApiKeyEntity>, i64)>;
    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        input: UpdateApiKey,
    ) -> RepositoryFuture<'a, ()>;
    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        by: Option<i64>,
    ) -> RepositoryFuture<'a, ()>;
}

#[derive(Clone, PartialEq, Eq)]
pub struct CreateApiKey {
    pub user_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub hash: String,
    pub redacted: String,
    pub by: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct ApiKeyFilter {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
    pub user_id: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateApiKey {
    pub name: Option<String>,
    pub description: Option<String>,
    pub preferences: Option<Value>,
    pub by: Option<i64>,
}
