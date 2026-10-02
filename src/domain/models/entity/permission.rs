use serde::{Deserialize, Serialize};
use serde_json::Value;

use super::AuditCreateUpdateDelete;

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Permission {
    pub id: i64,
    pub name: String,
    pub description: String,
    pub preferences: Value,
    #[serde(flatten)]
    pub audit: AuditCreateUpdateDelete,
}
