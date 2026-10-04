use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::domain::models::{ApiKey, AuditCreateUpdateDelete};

/// Management data without credential hashes; never use for authentication.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CachedApiKey {
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

impl From<ApiKey> for CachedApiKey {
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
