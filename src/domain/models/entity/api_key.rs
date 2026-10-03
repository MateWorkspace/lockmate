use serde::{Deserialize, Serialize};
use serde_json::Value;

use super::AuditCreateUpdateDelete;

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct ApiKey {
    pub id: i64,
    #[serde(default, skip_serializing_if = "is_zero")]
    pub user_id: i64,
    pub name: String,
    pub description: String,
    pub hash: String,
    pub redacted: String,
    pub preferences: Value,
    #[serde(flatten)]
    pub audit: AuditCreateUpdateDelete,
}

fn is_zero(value: &i64) -> bool {
    *value == 0
}

pub struct GeneratedApiKey {
    pub raw: String,
    pub hash: String,
    pub redacted: String,
}
