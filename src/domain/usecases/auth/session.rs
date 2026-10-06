use crate::domain::{
    models::{AccessClaims, AppContext, Permission, Role, Space, SpaceMember},
    usecases::{UsecaseFuture, shared::UserView},
};

/// Authentication within an existing space membership, using live repository data.
pub trait Session: Send + Sync {
    fn login<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        input: LoginRequest,
    ) -> UsecaseFuture<'a, SessionResult>;

    fn refresh<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        input: RefreshRequest,
    ) -> UsecaseFuture<'a, SessionResult>;

    /// Validate the token and resolve current identity, membership, and effective grants.
    fn authenticate_access_token<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        access_token: &'a str,
    ) -> UsecaseFuture<'a, AccessClaims>;

    fn authenticate_api_key<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        api_key: &'a str,
    ) -> UsecaseFuture<'a, AccessClaims>;
}

pub struct LoginRequest {
    pub username: String,
    pub password: String,
}

pub struct RefreshRequest {
    pub refresh_token: String,
}

pub struct SessionResult {
    pub user: UserView,
    pub space: Space,
    pub member: SpaceMember,
    pub roles: Vec<Role>,
    pub permissions: Vec<Permission>,
    pub access_token: String,
    pub refresh_token: String,
}
