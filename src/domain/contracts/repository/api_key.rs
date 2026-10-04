use serde_json::Value;

use crate::domain::models::{ApiKey as ApiKeyEntity, AppContext, Space, SpaceMember, User};

use super::RepositoryFuture;

pub trait ApiKey: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        input: CreateApiKey,
    ) -> RepositoryFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> RepositoryFuture<'a, ApiKeyEntity>;

    fn read_by_hash<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        hash: &'a str,
    ) -> RepositoryFuture<'a, ApiKeyEntity>;

    /// Resolves a key with live parents and active space/membership.
    /// Unavailable credentials return ApiKeyNotFound.
    fn read_active_by_hash<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        hash: &'a str,
    ) -> RepositoryFuture<'a, ApiKeyDetails>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: ApiKeyFilter,
    ) -> RepositoryFuture<'a, (Vec<ApiKeyEntity>, i64)>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
        input: UpdateApiKey,
    ) -> RepositoryFuture<'a, ()>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
        by: Option<i64>,
    ) -> RepositoryFuture<'a, ()>;
}

#[derive(Clone, PartialEq, Eq)]
pub struct CreateApiKey {
    pub member_id: i64,
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
    pub member_id: Option<i64>,
    pub user_id: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateApiKey {
    pub name: Option<String>,
    pub description: Option<String>,
    pub preferences: Option<Value>,
    pub by: Option<i64>,
}

#[derive(Clone, PartialEq)]
pub struct ApiKeyDetails {
    pub api_key: ApiKeyEntity,
    pub member: SpaceMember,
    pub user: User,
    pub space: Space,
}
