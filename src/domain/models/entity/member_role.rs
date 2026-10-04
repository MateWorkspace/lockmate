use serde::{Deserialize, Serialize};

use super::AuditCreate;

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct MemberRole {
    pub id: i64,
    pub space_id: i64,
    pub member_id: i64,
    pub role_id: i64,
    #[serde(flatten)]
    pub audit: AuditCreate,
}
