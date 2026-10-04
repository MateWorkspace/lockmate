use serde_json::Value;

use crate::domain::models::{AppContext, SpaceMember as SpaceMemberEntity, User};

use super::RepositoryFuture;

pub trait SpaceMember: Send + Sync {
    /// Creates the membership and assigns the current default role atomically.
    /// Returns RoleNotFound without creating records if the space has no default.
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        input: CreateSpaceMember,
    ) -> RepositoryFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> RepositoryFuture<'a, SpaceMemberEntity>;

    fn read_by_user_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        user_id: i64,
    ) -> RepositoryFuture<'a, SpaceMemberEntity>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: SpaceMemberFilter,
    ) -> RepositoryFuture<'a, (Vec<SpaceMemberWithUser>, i64)>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
        input: UpdateSpaceMember,
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
pub struct CreateSpaceMember {
    pub user_id: i64,
    pub is_active: Option<bool>,
    pub by: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct SpaceMemberFilter {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
    pub user_id: Option<i64>,
    pub role_id: Option<i64>,
    pub is_active: Option<bool>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateSpaceMember {
    pub is_active: Option<bool>,
    pub preferences: Option<Value>,
    pub by: Option<i64>,
}

#[derive(Clone, PartialEq)]
pub struct SpaceMemberWithUser {
    pub member: SpaceMemberEntity,
    pub user: User,
}
