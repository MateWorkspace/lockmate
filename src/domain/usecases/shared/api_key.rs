use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::domain::models::{ApiKey, AuditCreateUpdateDelete};

/// Public usecase data without credential hashes.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct ApiKeyView {
    pub id: i64,
    pub space_id: i64,
    pub member_id: i64,
    pub name: String,
    pub description: String,
    pub redacted: String,
    pub preferences: Value,
    #[serde(flatten)]
    pub audit: AuditCreateUpdateDelete,
}

impl From<ApiKey> for ApiKeyView {
    fn from(value: ApiKey) -> Self {
        Self {
            id: value.id,
            space_id: value.space_id,
            member_id: value.member_id,
            name: value.name,
            description: value.description,
            redacted: value.redacted,
            preferences: value.preferences,
            audit: value.audit,
        }
    }
}

impl From<crate::domain::models::CachedApiKey> for ApiKeyView {
    fn from(value: crate::domain::models::CachedApiKey) -> Self {
        Self {
            id: value.id,
            space_id: value.space_id,
            member_id: value.member_id,
            name: value.name,
            description: value.description,
            redacted: value.redacted,
            preferences: value.preferences,
            audit: value.audit,
        }
    }
}

/// The raw key is returned once, after successful creation.
pub struct CreatedApiKey {
    pub api_key: ApiKeyView,
    pub raw: String,
}
