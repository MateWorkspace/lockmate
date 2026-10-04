use crate::domain::models::{CacheFamily, CacheQuery, CachingError};
use serde_json::{Value, json};

#[derive(Clone, Copy, PartialEq, Eq)]
pub(crate) enum PayloadKind {
    SpaceRecord,
    SpacePage,
    UserRecord,
    UserPage,
    PermissionRecord,
    PermissionPage,
    RoleRecord,
    RolePage,
    SpaceMemberRecord,
    SpaceMemberPage,
    RolePermissionDetails,
    RolePermissionWithPermissionList,
    RolePermissionWithRoleList,
    MemberRoleDetails,
    MemberRoleWithRoleList,
    MemberRoleWithMemberList,
    ApiKeyRecord,
    ApiKeyPage,
}

pub(crate) struct QueryInfo {
    pub scope: String,
    pub entity: &'static str,
    pub operation: &'static str,
    pub selector: Value,
    pub families: Vec<CacheFamily>,
    pub kind: PayloadKind,
}

pub(crate) fn positive(id: i64) -> Result<(), CachingError> {
    if id <= 0 {
        return Err(CachingError::BadArgs);
    }
    Ok(())
}

pub(crate) fn family(family: CacheFamily) -> Result<(String, &'static str), CachingError> {
    let (id, entity) = match family {
        CacheFamily::Users => return Ok(("global".into(), "user")),
        CacheFamily::Spaces => return Ok(("global".into(), "space")),
        CacheFamily::Permissions { space_id } => (space_id, "permission"),
        CacheFamily::Roles { space_id } => (space_id, "role"),
        CacheFamily::SpaceMembers { space_id } => (space_id, "space_member"),
        CacheFamily::RolePermissions { space_id } => (space_id, "role_permission"),
        CacheFamily::MemberRoles { space_id } => (space_id, "member_role"),
        CacheFamily::ApiKeys { space_id } => (space_id, "api_key"),
    };
    positive(id)?;
    Ok((format!("space:{id}"), entity))
}

pub(crate) fn describe(query: &CacheQuery) -> Result<QueryInfo, CachingError> {
    let mut info = match query {
        CacheQuery::SpaceById { id } => {
            positive(*id)?;
            QueryInfo {
                scope: "global".to_owned(),
                entity: "space",
                operation: "read_by_id",
                selector: json!([json!(id)]),
                families: vec![CacheFamily::Spaces],
                kind: PayloadKind::SpaceRecord,
            }
        }
        CacheQuery::SpaceBySlug { slug } => {
            if slug.is_empty() {
                return Err(CachingError::BadArgs);
            }
            QueryInfo {
                scope: "global".to_owned(),
                entity: "space",
                operation: "read_by_slug",
                selector: json!([json!(slug)]),
                families: vec![CacheFamily::Spaces],
                kind: PayloadKind::SpaceRecord,
            }
        }
        CacheQuery::SpaceByFilter { filter } => QueryInfo {
            scope: "global".to_owned(),
            entity: "space",
            operation: "read_by_filter",
            selector: json!([json!([
                filter.page,
                filter.limit,
                filter.search,
                filter.is_active
            ])]),
            families: vec![CacheFamily::Spaces],
            kind: PayloadKind::SpacePage,
        },
        CacheQuery::UserById { id } => {
            positive(*id)?;
            QueryInfo {
                scope: "global".to_owned(),
                entity: "user",
                operation: "read_by_id",
                selector: json!([json!(id)]),
                families: vec![CacheFamily::Users],
                kind: PayloadKind::UserRecord,
            }
        }
        CacheQuery::UserByUsername { username } => {
            if username.is_empty() {
                return Err(CachingError::BadArgs);
            }
            QueryInfo {
                scope: "global".to_owned(),
                entity: "user",
                operation: "read_by_username",
                selector: json!([json!(username)]),
                families: vec![CacheFamily::Users],
                kind: PayloadKind::UserRecord,
            }
        }
        CacheQuery::UserByEmail { email } => {
            if email.is_empty() {
                return Err(CachingError::BadArgs);
            }
            QueryInfo {
                scope: "global".to_owned(),
                entity: "user",
                operation: "read_by_email",
                selector: json!([json!(email)]),
                families: vec![CacheFamily::Users],
                kind: PayloadKind::UserRecord,
            }
        }
        CacheQuery::UserByPhone { phone } => {
            if phone.is_empty() {
                return Err(CachingError::BadArgs);
            }
            QueryInfo {
                scope: "global".to_owned(),
                entity: "user",
                operation: "read_by_phone",
                selector: json!([json!(phone)]),
                families: vec![CacheFamily::Users],
                kind: PayloadKind::UserRecord,
            }
        }
        CacheQuery::UserByFilter { filter } => QueryInfo {
            scope: "global".to_owned(),
            entity: "user",
            operation: "read_by_filter",
            selector: json!([json!([
                filter.page,
                filter.limit,
                filter.search,
                filter.is_email_verified,
                filter.is_phone_verified
            ])]),
            families: vec![CacheFamily::Users],
            kind: PayloadKind::UserPage,
        },
        CacheQuery::PermissionById { space_id, id } => {
            positive(*space_id)?;
            positive(*id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "permission",
                operation: "read_by_id",
                selector: json!([json!(space_id), json!(id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Permissions {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::PermissionRecord,
            }
        }
        CacheQuery::PermissionBySlug { space_id, slug } => {
            positive(*space_id)?;
            if slug.is_empty() {
                return Err(CachingError::BadArgs);
            }
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "permission",
                operation: "read_by_slug",
                selector: json!([json!(space_id), json!(slug)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Permissions {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::PermissionRecord,
            }
        }
        CacheQuery::PermissionByFilter { space_id, filter } => {
            positive(*space_id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "permission",
                operation: "read_by_filter",
                selector: json!([
                    json!(space_id),
                    json!([filter.page, filter.limit, filter.search])
                ]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Permissions {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::PermissionPage,
            }
        }
        CacheQuery::RoleById { space_id, id } => {
            positive(*space_id)?;
            positive(*id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "role",
                operation: "read_by_id",
                selector: json!([json!(space_id), json!(id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::RoleRecord,
            }
        }
        CacheQuery::RoleBySlug { space_id, slug } => {
            positive(*space_id)?;
            if slug.is_empty() {
                return Err(CachingError::BadArgs);
            }
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "role",
                operation: "read_by_slug",
                selector: json!([json!(space_id), json!(slug)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::RoleRecord,
            }
        }
        CacheQuery::RoleDefault { space_id } => {
            positive(*space_id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "role",
                operation: "read_default",
                selector: json!([json!(space_id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::RoleRecord,
            }
        }
        CacheQuery::RoleByFilter { space_id, filter } => {
            positive(*space_id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "role",
                operation: "read_by_filter",
                selector: json!([
                    json!(space_id),
                    json!([filter.page, filter.limit, filter.search, filter.is_default])
                ]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::RolePage,
            }
        }
        CacheQuery::SpaceMemberById { space_id, id } => {
            positive(*space_id)?;
            positive(*id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "space_member",
                operation: "read_by_id",
                selector: json!([json!(space_id), json!(id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Users,
                    CacheFamily::SpaceMembers {
                        space_id: *space_id,
                    },
                    CacheFamily::MemberRoles {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::SpaceMemberRecord,
            }
        }
        CacheQuery::SpaceMemberByUserId { space_id, user_id } => {
            positive(*space_id)?;
            positive(*user_id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "space_member",
                operation: "read_by_user_id",
                selector: json!([json!(space_id), json!(user_id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Users,
                    CacheFamily::SpaceMembers {
                        space_id: *space_id,
                    },
                    CacheFamily::MemberRoles {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::SpaceMemberRecord,
            }
        }
        CacheQuery::SpaceMemberByFilter { space_id, filter } => {
            positive(*space_id)?;
            if let Some(id) = filter.user_id {
                positive(id)?;
            }
            if let Some(id) = filter.role_id {
                positive(id)?;
            }
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "space_member",
                operation: "read_by_filter",
                selector: json!([
                    json!(space_id),
                    json!([
                        filter.page,
                        filter.limit,
                        filter.search,
                        filter.user_id,
                        filter.role_id,
                        filter.is_active
                    ])
                ]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Users,
                    CacheFamily::SpaceMembers {
                        space_id: *space_id,
                    },
                    CacheFamily::MemberRoles {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::SpaceMemberPage,
            }
        }
        CacheQuery::RolePermissionById { space_id, id } => {
            positive(*space_id)?;
            positive(*id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "role_permission",
                operation: "read_by_id",
                selector: json!([json!(space_id), json!(id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::RolePermissions {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                    CacheFamily::Permissions {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::RolePermissionDetails,
            }
        }
        CacheQuery::RolePermissionByPair {
            space_id,
            role_id,
            permission_id,
        } => {
            positive(*space_id)?;
            positive(*role_id)?;
            positive(*permission_id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "role_permission",
                operation: "read_by_role_id_and_permission_id",
                selector: json!([json!(space_id), json!(role_id), json!(permission_id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::RolePermissions {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                    CacheFamily::Permissions {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::RolePermissionDetails,
            }
        }
        CacheQuery::RolePermissionByRoleId { space_id, role_id } => {
            positive(*space_id)?;
            positive(*role_id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "role_permission",
                operation: "read_by_role_id",
                selector: json!([json!(space_id), json!(role_id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::RolePermissions {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                    CacheFamily::Permissions {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::RolePermissionWithPermissionList,
            }
        }
        CacheQuery::RolePermissionByPermissionId {
            space_id,
            permission_id,
        } => {
            positive(*space_id)?;
            positive(*permission_id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "role_permission",
                operation: "read_by_permission_id",
                selector: json!([json!(space_id), json!(permission_id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::RolePermissions {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                    CacheFamily::Permissions {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::RolePermissionWithRoleList,
            }
        }
        CacheQuery::MemberRoleById { space_id, id } => {
            positive(*space_id)?;
            positive(*id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "member_role",
                operation: "read_by_id",
                selector: json!([json!(space_id), json!(id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Users,
                    CacheFamily::MemberRoles {
                        space_id: *space_id,
                    },
                    CacheFamily::SpaceMembers {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::MemberRoleDetails,
            }
        }
        CacheQuery::MemberRoleByPair {
            space_id,
            member_id,
            role_id,
        } => {
            positive(*space_id)?;
            positive(*member_id)?;
            positive(*role_id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "member_role",
                operation: "read_by_member_id_and_role_id",
                selector: json!([json!(space_id), json!(member_id), json!(role_id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Users,
                    CacheFamily::MemberRoles {
                        space_id: *space_id,
                    },
                    CacheFamily::SpaceMembers {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::MemberRoleDetails,
            }
        }
        CacheQuery::MemberRoleByMemberId {
            space_id,
            member_id,
        } => {
            positive(*space_id)?;
            positive(*member_id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "member_role",
                operation: "read_by_member_id",
                selector: json!([json!(space_id), json!(member_id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Users,
                    CacheFamily::MemberRoles {
                        space_id: *space_id,
                    },
                    CacheFamily::SpaceMembers {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::MemberRoleWithRoleList,
            }
        }
        CacheQuery::MemberRoleByRoleId { space_id, role_id } => {
            positive(*space_id)?;
            positive(*role_id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "member_role",
                operation: "read_by_role_id",
                selector: json!([json!(space_id), json!(role_id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Users,
                    CacheFamily::MemberRoles {
                        space_id: *space_id,
                    },
                    CacheFamily::SpaceMembers {
                        space_id: *space_id,
                    },
                    CacheFamily::Roles {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::MemberRoleWithMemberList,
            }
        }
        CacheQuery::ApiKeyById { space_id, id } => {
            positive(*space_id)?;
            positive(*id)?;
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "api_key",
                operation: "read_by_id",
                selector: json!([json!(space_id), json!(id)]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Users,
                    CacheFamily::ApiKeys {
                        space_id: *space_id,
                    },
                    CacheFamily::SpaceMembers {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::ApiKeyRecord,
            }
        }
        CacheQuery::ApiKeyByFilter { space_id, filter } => {
            positive(*space_id)?;
            if let Some(id) = filter.member_id {
                positive(id)?;
            }
            if let Some(id) = filter.user_id {
                positive(id)?;
            }
            QueryInfo {
                scope: format!("space:{}", space_id),
                entity: "api_key",
                operation: "read_by_filter",
                selector: json!([
                    json!(space_id),
                    json!([
                        filter.page,
                        filter.limit,
                        filter.search,
                        filter.member_id,
                        filter.user_id
                    ])
                ]),
                families: vec![
                    CacheFamily::Spaces,
                    CacheFamily::Users,
                    CacheFamily::ApiKeys {
                        space_id: *space_id,
                    },
                    CacheFamily::SpaceMembers {
                        space_id: *space_id,
                    },
                ],
                kind: PayloadKind::ApiKeyPage,
            }
        }
    };
    info.families.sort();
    Ok(info)
}
