use serde_json::Value;

use crate::domain::models::{AppContext, Space as SpaceEntity};

use super::RepositoryFuture;

pub trait Space: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        input: CreateSpace,
    ) -> RepositoryFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> RepositoryFuture<'a, SpaceEntity>;

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        slug: &'a str,
    ) -> RepositoryFuture<'a, SpaceEntity>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: SpaceFilter,
    ) -> RepositoryFuture<'a, (Vec<SpaceEntity>, i64)>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        input: UpdateSpace,
    ) -> RepositoryFuture<'a, ()>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        by: Option<i64>,
    ) -> RepositoryFuture<'a, ()>;
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CreateSpace {
    pub slug: String,
    pub name: String,
    pub description: Option<String>,
    pub is_active: Option<bool>,
    pub by: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct SpaceFilter {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
    pub is_active: Option<bool>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateSpace {
    pub name: Option<String>,
    pub description: Option<String>,
    pub is_active: Option<bool>,
    pub preferences: Option<Value>,
    pub by: Option<i64>,
}
