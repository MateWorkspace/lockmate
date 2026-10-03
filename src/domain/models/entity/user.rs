use serde::{Deserialize, Serialize};
use serde_json::Value;

use super::AuditCreateUpdateDelete;

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct User {
    pub id: i64,
    pub role_id: i64,
    pub name: String,
    pub bio: String,
    pub username: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub email: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub phone: Option<String>,
    pub password_hash: String,
    pub is_email_verified: bool,
    pub is_phone_verified: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub avatar_path: Option<String>,
    pub preferences: Value,
    #[serde(flatten)]
    pub audit: AuditCreateUpdateDelete,
}
