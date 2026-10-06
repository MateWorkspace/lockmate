use serde::{Deserialize, Serialize};

use crate::domain::models::{MemberRole, Role, SpaceMember};

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct MemberRoleDetails {
    pub member_role: MemberRole,
    pub member: SpaceMember,
    pub role: Role,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct MemberRoleWithMember {
    pub member_role: MemberRole,
    pub member: SpaceMember,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct MemberRoleWithRole {
    pub member_role: MemberRole,
    pub role: Role,
}

impl From<crate::domain::contracts::repository::MemberRoleDetails> for MemberRoleDetails {
    fn from(value: crate::domain::contracts::repository::MemberRoleDetails) -> Self {
        Self {
            member_role: value.member_role,
            member: value.member,
            role: value.role,
        }
    }
}

impl From<crate::domain::models::CachedMemberRoleDetails> for MemberRoleDetails {
    fn from(value: crate::domain::models::CachedMemberRoleDetails) -> Self {
        Self {
            member_role: value.member_role,
            member: value.member,
            role: value.role,
        }
    }
}

impl From<crate::domain::contracts::repository::MemberRoleWithMember> for MemberRoleWithMember {
    fn from(value: crate::domain::contracts::repository::MemberRoleWithMember) -> Self {
        Self {
            member_role: value.member_role,
            member: value.member,
        }
    }
}

impl From<crate::domain::models::CachedMemberRoleWithMember> for MemberRoleWithMember {
    fn from(value: crate::domain::models::CachedMemberRoleWithMember) -> Self {
        Self {
            member_role: value.member_role,
            member: value.member,
        }
    }
}

impl From<crate::domain::contracts::repository::MemberRoleWithRole> for MemberRoleWithRole {
    fn from(value: crate::domain::contracts::repository::MemberRoleWithRole) -> Self {
        Self {
            member_role: value.member_role,
            role: value.role,
        }
    }
}

impl From<crate::domain::models::CachedMemberRoleWithRole> for MemberRoleWithRole {
    fn from(value: crate::domain::models::CachedMemberRoleWithRole) -> Self {
        Self {
            member_role: value.member_role,
            role: value.role,
        }
    }
}
