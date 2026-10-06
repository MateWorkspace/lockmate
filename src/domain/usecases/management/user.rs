use serde_json::Value;

use crate::domain::models::{AccessClaims, AppContext};
use crate::domain::usecases::{
    UsecaseFuture,
    shared::{Page, UserView},
};

pub trait User: Send + Sync {
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
    ) -> UsecaseFuture<'a, UserView>;

    fn read_by_username<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        username: &'a str,
    ) -> UsecaseFuture<'a, UserView>;

    fn read_by_email<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        email: &'a str,
    ) -> UsecaseFuture<'a, UserView>;

    fn read_by_phone<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        phone: &'a str,
    ) -> UsecaseFuture<'a, UserView>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        filter: FilterRequest,
    ) -> UsecaseFuture<'a, Page<UserView>>;

    fn update_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        id: i64,
        input: UpdateRequest,
    ) -> UsecaseFuture<'a, ()>;

    fn reset_password<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        id: i64,
        input: ResetPasswordRequest,
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
    pub bio: Option<String>,
    pub username: String,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub password: String,
    pub is_email_verified: Option<bool>,
    pub is_phone_verified: Option<bool>,
    pub avatar_path: Option<String>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct FilterRequest {
    pub page: i64,
    pub limit: i64,
    pub search: Option<String>,
    pub is_email_verified: Option<bool>,
    pub is_phone_verified: Option<bool>,
}

#[derive(Clone, Default, PartialEq)]
pub struct UpdateRequest {
    pub name: Option<String>,
    pub bio: Option<String>,
    pub username: Option<String>,
    pub email: Option<Option<String>>,
    pub phone: Option<Option<String>>,
    pub is_email_verified: Option<bool>,
    pub is_phone_verified: Option<bool>,
    pub avatar_path: Option<Option<String>>,
    pub preferences: Option<Value>,
}

pub struct ResetPasswordRequest {
    pub password: String,
}
