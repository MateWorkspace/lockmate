use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::domain::models::{AuditCreateUpdateDelete, User};

/// Public usecase data without credential hashes.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct UserView {
    pub id: i64,
    pub name: String,
    pub bio: String,
    pub username: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub email: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub phone: Option<String>,
    pub is_email_verified: bool,
    pub is_phone_verified: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub avatar_path: Option<String>,
    pub preferences: Value,
    #[serde(flatten)]
    pub audit: AuditCreateUpdateDelete,
}

impl From<User> for UserView {
    fn from(value: User) -> Self {
        Self {
            id: value.id,
            name: value.name,
            bio: value.bio,
            username: value.username,
            email: value.email,
            phone: value.phone,
            is_email_verified: value.is_email_verified,
            is_phone_verified: value.is_phone_verified,
            avatar_path: value.avatar_path,
            preferences: value.preferences,
            audit: value.audit,
        }
    }
}

impl From<crate::domain::models::CachedUser> for UserView {
    fn from(value: crate::domain::models::CachedUser) -> Self {
        Self {
            id: value.id,
            name: value.name,
            bio: value.bio,
            username: value.username,
            email: value.email,
            phone: value.phone,
            is_email_verified: value.is_email_verified,
            is_phone_verified: value.is_phone_verified,
            avatar_path: value.avatar_path,
            preferences: value.preferences,
            audit: value.audit,
        }
    }
}
