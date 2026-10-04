use serde::{Deserialize, Serialize};

use crate::domain::models::{MemberRole, Role, SpaceMember};

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CachedMemberRoleDetails {
    pub member_role: MemberRole,
    pub member: SpaceMember,
    pub role: Role,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CachedMemberRoleWithMember {
    pub member_role: MemberRole,
    pub member: SpaceMember,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CachedMemberRoleWithRole {
    pub member_role: MemberRole,
    pub role: Role,
}
