use super::query::{self, PayloadKind};
use crate::domain::models::{CacheQuery, CachingError};
use serde_json::Value;
use std::time::Duration;

pub(crate) fn ttl(ttl: Duration) -> Result<i64, CachingError> {
    if ttl.is_zero() {
        return Err(CachingError::BadArgs);
    }
    let millis = ttl.as_nanos().div_ceil(1_000_000);
    i64::try_from(millis).map_err(|_| CachingError::BadArgs)
}

fn require(condition: bool) -> Result<(), CachingError> {
    if condition {
        Ok(())
    } else {
        Err(CachingError::BadArgs)
    }
}

fn id(value: &Value, name: &str) -> Result<i64, CachingError> {
    let id = value[name].as_i64().ok_or(CachingError::BadArgs)?;
    query::positive(id)?;
    Ok(id)
}

fn equal_id(value: &Value, name: &str, expected: i64) -> Result<(), CachingError> {
    require(id(value, name)? == expected)
}

fn equal_text(value: &Value, name: &str, expected: &str) -> Result<(), CachingError> {
    require(value[name].as_str() == Some(expected))
}

fn flag(value: &Value, name: &str, expected: Option<bool>) -> Result<(), CachingError> {
    require(expected.is_none_or(|expected| value[name].as_bool() == Some(expected)))
}

fn optional_id(value: &Value, name: &str, expected: Option<i64>) -> Result<(), CachingError> {
    if let Some(expected) = expected {
        equal_id(value, name, expected)?;
    }
    Ok(())
}

fn record(value: &Value, entity: &str, space: Option<i64>) -> Result<(), CachingError> {
    id(value, "id")?;
    require(value.get("deleted_at").is_none_or(Value::is_null))?;
    if let Some(space) = space {
        equal_id(value, "space_id", space)?;
    }
    match entity {
        "user" => require(value.get("password_hash").is_none())?,
        "space_member" => {
            id(value, "user_id")?;
        }
        "api_key" => {
            id(value, "member_id")?;
            require(value.get("hash").is_none())?;
        }
        "member_role" => {
            id(value, "member_id")?;
            id(value, "role_id")?;
        }
        "role_permission" => {
            id(value, "role_id")?;
            id(value, "permission_id")?;
        }
        _ => {}
    }
    Ok(())
}

fn member_with_user(value: &Value, space: i64) -> Result<(), CachingError> {
    record(&value["member"], "space_member", Some(space))?;
    record(&value["user"], "user", None)?;
    equal_id(&value["member"], "user_id", id(&value["user"], "id")?)
}

fn assignment(
    value: &Value,
    entity: &str,
    space: i64,
    with_member: bool,
    with_role: bool,
    with_permission: bool,
) -> Result<(), CachingError> {
    record(&value[entity], entity, Some(space))?;
    if with_member {
        record(&value["member"], "space_member", Some(space))?;
        equal_id(&value[entity], "member_id", id(&value["member"], "id")?)?;
    }
    if with_role {
        record(&value["role"], "role", Some(space))?;
        equal_id(&value[entity], "role_id", id(&value["role"], "id")?)?;
    }
    if with_permission {
        record(&value["permission"], "permission", Some(space))?;
        equal_id(
            &value[entity],
            "permission_id",
            id(&value["permission"], "id")?,
        )?;
    }
    Ok(())
}

fn page(
    value: &Value,
    limit: i64,
    mut validate: impl FnMut(&Value) -> Result<(), CachingError>,
) -> Result<(), CachingError> {
    require(value["total"].as_i64().is_some_and(|total| total >= 0))?;
    let items = value["items"].as_array().ok_or(CachingError::BadArgs)?;
    require(items.len() as u128 <= limit.max(0) as u128)?;
    for item in items {
        validate(item)?;
    }
    Ok(())
}

fn list(
    value: &Value,
    mut validate: impl FnMut(&Value) -> Result<(), CachingError>,
) -> Result<(), CachingError> {
    for item in value.as_array().ok_or(CachingError::BadArgs)? {
        validate(item)?;
    }
    Ok(())
}

pub(crate) fn payload(
    query: &CacheQuery,
    kind: PayloadKind,
    value: &Value,
) -> Result<(), CachingError> {
    require(query::describe(query)?.kind == kind)?;
    match query {
        CacheQuery::SpaceById { id } => {
            record(value, "space", None)?;
            equal_id(value, "id", *id)
        }
        CacheQuery::SpaceBySlug { slug } => {
            record(value, "space", None)?;
            equal_text(value, "slug", slug)
        }
        CacheQuery::UserById { id } => {
            record(value, "user", None)?;
            equal_id(value, "id", *id)
        }
        CacheQuery::UserByUsername { username } => {
            record(value, "user", None)?;
            equal_text(value, "username", username)
        }
        CacheQuery::UserByEmail { email } => {
            record(value, "user", None)?;
            equal_text(value, "email", email)
        }
        CacheQuery::UserByPhone { phone } => {
            record(value, "user", None)?;
            equal_text(value, "phone", phone)
        }
        CacheQuery::PermissionById { space_id, id } => {
            record(value, "permission", Some(*space_id))?;
            equal_id(value, "id", *id)
        }
        CacheQuery::PermissionBySlug { space_id, slug } => {
            record(value, "permission", Some(*space_id))?;
            equal_text(value, "slug", slug)
        }
        CacheQuery::RoleById { space_id, id } => {
            record(value, "role", Some(*space_id))?;
            equal_id(value, "id", *id)
        }
        CacheQuery::RoleBySlug { space_id, slug } => {
            record(value, "role", Some(*space_id))?;
            equal_text(value, "slug", slug)
        }
        CacheQuery::RoleDefault { space_id } => {
            record(value, "role", Some(*space_id))?;
            flag(value, "is_default", Some(true))
        }
        CacheQuery::SpaceMemberById { space_id, id } => {
            record(value, "space_member", Some(*space_id))?;
            equal_id(value, "id", *id)
        }
        CacheQuery::SpaceMemberByUserId { space_id, user_id } => {
            record(value, "space_member", Some(*space_id))?;
            equal_id(value, "user_id", *user_id)
        }
        CacheQuery::ApiKeyById { space_id, id } => {
            record(value, "api_key", Some(*space_id))?;
            equal_id(value, "id", *id)
        }
        CacheQuery::SpaceByFilter { filter } => page(value, filter.limit, |v| {
            record(v, "space", None)?;
            flag(v, "is_active", filter.is_active)
        }),
        CacheQuery::UserByFilter { filter } => page(value, filter.limit, |v| {
            record(v, "user", None)?;
            flag(v, "is_email_verified", filter.is_email_verified)?;
            flag(v, "is_phone_verified", filter.is_phone_verified)
        }),
        CacheQuery::PermissionByFilter { space_id, filter } => page(value, filter.limit, |v| {
            record(v, "permission", Some(*space_id))
        }),
        CacheQuery::RoleByFilter { space_id, filter } => page(value, filter.limit, |v| {
            record(v, "role", Some(*space_id))?;
            flag(v, "is_default", filter.is_default)
        }),
        CacheQuery::SpaceMemberByFilter { space_id, filter } => page(value, filter.limit, |v| {
            member_with_user(v, *space_id)?;
            optional_id(&v["member"], "user_id", filter.user_id)?;
            flag(&v["member"], "is_active", filter.is_active)
        }),
        CacheQuery::ApiKeyByFilter { space_id, filter } => page(value, filter.limit, |v| {
            record(v, "api_key", Some(*space_id))?;
            optional_id(v, "member_id", filter.member_id)
        }),
        CacheQuery::RolePermissionById { space_id, id } => {
            assignment(value, "role_permission", *space_id, false, true, true)?;
            equal_id(&value["role_permission"], "id", *id)
        }
        CacheQuery::RolePermissionByPair {
            space_id,
            role_id,
            permission_id,
        } => {
            assignment(value, "role_permission", *space_id, false, true, true)?;
            equal_id(&value["role_permission"], "role_id", *role_id)?;
            equal_id(&value["role_permission"], "permission_id", *permission_id)
        }
        CacheQuery::MemberRoleById { space_id, id } => {
            assignment(value, "member_role", *space_id, true, true, false)?;
            equal_id(&value["member_role"], "id", *id)
        }
        CacheQuery::MemberRoleByPair {
            space_id,
            member_id,
            role_id,
        } => {
            assignment(value, "member_role", *space_id, true, true, false)?;
            equal_id(&value["member_role"], "member_id", *member_id)?;
            equal_id(&value["member_role"], "role_id", *role_id)
        }
        CacheQuery::RolePermissionByRoleId { space_id, role_id } => list(value, |v| {
            assignment(v, "role_permission", *space_id, false, false, true)?;
            equal_id(&v["role_permission"], "role_id", *role_id)
        }),
        CacheQuery::RolePermissionByPermissionId {
            space_id,
            permission_id,
        } => list(value, |v| {
            assignment(v, "role_permission", *space_id, false, true, false)?;
            equal_id(&v["role_permission"], "permission_id", *permission_id)
        }),
        CacheQuery::MemberRoleByMemberId {
            space_id,
            member_id,
        } => list(value, |v| {
            assignment(v, "member_role", *space_id, false, true, false)?;
            equal_id(&v["member_role"], "member_id", *member_id)
        }),
        CacheQuery::MemberRoleByRoleId { space_id, role_id } => list(value, |v| {
            assignment(v, "member_role", *space_id, true, false, false)?;
            equal_id(&v["member_role"], "role_id", *role_id)
        }),
    }
}
