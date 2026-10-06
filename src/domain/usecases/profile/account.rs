use serde_json::Value;

use crate::domain::{
    models::{AccessClaims, AppContext, Permission, Role, Space, SpaceMember},
    usecases::{UsecaseFuture, shared::UserView},
};

pub trait Account: Send + Sync {
    fn get_profile<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
    ) -> UsecaseFuture<'a, ProfileResult>;

    fn get_roles<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
    ) -> UsecaseFuture<'a, Vec<Role>>;

    fn get_permissions<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
    ) -> UsecaseFuture<'a, Vec<Permission>>;

    fn update_profile<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        input: UpdateProfileRequest,
    ) -> UsecaseFuture<'a, ()>;

    fn update_membership_preferences<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        input: UpdateMembershipPreferencesRequest,
    ) -> UsecaseFuture<'a, ()>;
}

#[derive(Debug, Clone, PartialEq)]
pub struct ProfileResult {
    pub user: UserView,
    pub space: Space,
    pub member: SpaceMember,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct UpdateProfileRequest {
    pub name: Option<String>,
    pub bio: Option<String>,
    pub username: Option<String>,
    /// None leaves the email unchanged; Some(None) clears it.
    pub email: Option<Option<String>>,
    /// None leaves the phone unchanged; Some(None) clears it.
    pub phone: Option<Option<String>>,
    /// Metadata only; this operation does not upload or delete an object.
    pub avatar_path: Option<Option<String>>,
    pub preferences: Option<Value>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct UpdateMembershipPreferencesRequest {
    pub preferences: Value,
}
