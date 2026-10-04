use serde::{Deserialize, Serialize};
use serde_json::Value;

use super::AuditCreateUpdateDelete;

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct SpaceMember {
    pub id: i64,
    pub space_id: i64,
    pub user_id: i64,
    pub is_active: bool,
    pub preferences: Value,
    #[serde(flatten)]
    pub audit: AuditCreateUpdateDelete,
}
