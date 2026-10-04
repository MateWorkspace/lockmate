use serde_json::Value;

use crate::domain::models::{AppContext, User as UserEntity};

use super::RepositoryFuture;

pub trait User: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        input: CreateUser,
    ) -> RepositoryFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> RepositoryFuture<'a, UserEntity>;

    fn read_by_username<'a>(
        &'a self,
        context: &'a AppContext,
        username: &'a str,
    ) -> RepositoryFuture<'a, UserEntity>;

    fn read_by_email<'a>(
        &'a self,
        context: &'a AppContext,
        email: &'a str,
    ) -> RepositoryFuture<'a, UserEntity>;

    fn read_by_phone<'a>(
        &'a self,
        context: &'a AppContext,
        phone: &'a str,
    ) -> RepositoryFuture<'a, UserEntity>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: UserFilter,
    ) -> RepositoryFuture<'a, (Vec<UserEntity>, i64)>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        input: UpdateUser,
    ) -> RepositoryFuture<'a, ()>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
        by: Option<i64>,
    ) -> RepositoryFuture<'a, ()>;
}

#[derive(Clone, PartialEq, Eq)]
pub struct CreateUser {
    pub name: String,
    pub bio: Option<String>,
    pub username: String,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub password_hash: String,
    pub is_email_verified: Option<bool>,
    pub is_phone_verified: Option<bool>,
    pub avatar_path: Option<String>,
    pub by: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct UserFilter {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
    pub is_email_verified: Option<bool>,
    pub is_phone_verified: Option<bool>,
}

#[derive(Clone, Default, PartialEq)]
pub struct UpdateUser {
    pub name: Option<String>,
    pub bio: Option<String>,
    pub username: Option<String>,
    pub email: Option<Option<String>>,
    pub phone: Option<Option<String>>,
    pub password_hash: Option<String>,
    pub is_email_verified: Option<bool>,
    pub is_phone_verified: Option<bool>,
    pub avatar_path: Option<Option<String>>,
    pub preferences: Option<Value>,
    pub by: Option<i64>,
}
