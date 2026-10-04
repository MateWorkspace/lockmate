use serde_json::Value;

use crate::domain::models::{AppContext, Permission as PermissionEntity};

use super::RepositoryFuture;

pub trait Permission: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        input: CreatePermission,
    ) -> RepositoryFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> RepositoryFuture<'a, PermissionEntity>;

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        slug: &'a str,
    ) -> RepositoryFuture<'a, PermissionEntity>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: PermissionFilter,
    ) -> RepositoryFuture<'a, (Vec<PermissionEntity>, i64)>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
        input: UpdatePermission,
    ) -> RepositoryFuture<'a, ()>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
        by: Option<i64>,
    ) -> RepositoryFuture<'a, ()>;
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CreatePermission {
    pub slug: String,
    pub name: String,
    pub description: Option<String>,
    pub by: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct PermissionFilter {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdatePermission {
    pub name: Option<String>,
    pub description: Option<String>,
    pub preferences: Option<Value>,
    pub by: Option<i64>,
}
