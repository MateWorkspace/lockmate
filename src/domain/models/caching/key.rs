use crate::domain::contracts::repository::{
    ApiKeyFilter, PermissionFilter, RoleFilter, SpaceFilter, SpaceMemberFilter, UserFilter,
};

use super::CacheRevision;

/// Complete lookup identity. No authentication or credential-hash selectors.
#[derive(Clone, PartialEq, Eq)]
pub enum CacheQuery {
    SpaceById {
        id: i64,
    },
    SpaceBySlug {
        slug: String,
    },
    SpaceByFilter {
        filter: SpaceFilter,
    },
    UserById {
        id: i64,
    },
    UserByUsername {
        username: String,
    },
    UserByEmail {
        email: String,
    },
    UserByPhone {
        phone: String,
    },
    UserByFilter {
        filter: UserFilter,
    },
    PermissionById {
        space_id: i64,
        id: i64,
    },
    PermissionBySlug {
        space_id: i64,
        slug: String,
    },
    PermissionByFilter {
        space_id: i64,
        filter: PermissionFilter,
    },
    RoleById {
        space_id: i64,
        id: i64,
    },
    RoleBySlug {
        space_id: i64,
        slug: String,
    },
    RoleDefault {
        space_id: i64,
    },
    RoleByFilter {
        space_id: i64,
        filter: RoleFilter,
    },
    SpaceMemberById {
        space_id: i64,
        id: i64,
    },
    SpaceMemberByUserId {
        space_id: i64,
        user_id: i64,
    },
    SpaceMemberByFilter {
        space_id: i64,
        filter: SpaceMemberFilter,
    },
    RolePermissionById {
        space_id: i64,
        id: i64,
    },
    RolePermissionByPair {
        space_id: i64,
        role_id: i64,
        permission_id: i64,
    },
    RolePermissionByRoleId {
        space_id: i64,
        role_id: i64,
    },
    RolePermissionByPermissionId {
        space_id: i64,
        permission_id: i64,
    },
    MemberRoleById {
        space_id: i64,
        id: i64,
    },
    MemberRoleByPair {
        space_id: i64,
        member_id: i64,
        role_id: i64,
    },
    MemberRoleByMemberId {
        space_id: i64,
        member_id: i64,
    },
    MemberRoleByRoleId {
        space_id: i64,
        role_id: i64,
    },
    ApiKeyById {
        space_id: i64,
        id: i64,
    },
    ApiKeyByFilter {
        space_id: i64,
        filter: ApiKeyFilter,
    },
}

/// Revisions must exactly cover the query dependencies, sorted by family.
#[derive(Clone, PartialEq, Eq)]
pub struct CacheEntryKey {
    pub query: CacheQuery,
    pub revisions: Vec<CacheRevision>,
}
