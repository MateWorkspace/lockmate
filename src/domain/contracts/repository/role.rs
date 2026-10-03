use serde_json::Value;

use crate::domain::models::{AppContext, Role as RoleEntity};

use super::RepositoryFuture;

pub trait Role: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        input: CreateRole,
    ) -> RepositoryFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> RepositoryFuture<'a, RoleEntity>;

    fn read_by_name<'a>(
        &'a self,
        context: &'a AppContext,
        name: &'a str,
    ) -> RepositoryFuture<'a, RoleEntity>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: RoleFilter,
    ) -> RepositoryFuture<'a, (Vec<RoleEntity>, i64)>;

    fn read_default<'a>(&'a self, context: &'a AppContext) -> RepositoryFuture<'a, RoleEntity>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        input: UpdateRole,
    ) -> RepositoryFuture<'a, ()>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        by: Option<i64>,
    ) -> RepositoryFuture<'a, ()>;
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CreateRole {
    pub name: String,
    pub description: Option<String>,
    pub is_default: Option<bool>,
    pub by: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct RoleFilter {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
    pub is_default: Option<bool>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateRole {
    pub name: Option<String>,
    pub description: Option<String>,
    pub is_default: Option<bool>,
    pub preferences: Option<Value>,
    pub by: Option<i64>,
}
