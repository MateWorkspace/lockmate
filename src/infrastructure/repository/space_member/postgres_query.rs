use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use super::super::shared::query::{pagination, search_pattern};
use crate::domain::{
    contracts::repository::{CreateSpaceMember, SpaceMemberFilter, UpdateSpaceMember},
    models::RepositoryError,
};

const COLUMNS: &str = "e.id AS m_id, e.space_id AS m_space_id, e.user_id AS m_user_id, e.is_active AS m_is_active, e.preferences AS m_preferences, e.created_at AS m_created_at, e.updated_at AS m_updated_at, e.deleted_at AS m_deleted_at, e.created_by AS m_created_by, e.updated_by AS m_updated_by, e.deleted_by AS m_deleted_by, u.id AS u_id, u.name AS u_name, u.bio AS u_bio, u.username AS u_username, u.email AS u_email, u.phone AS u_phone, u.password_hash AS u_password_hash, u.is_email_verified AS u_is_email_verified, u.is_phone_verified AS u_is_phone_verified, u.avatar_path AS u_avatar_path, u.preferences AS u_preferences, u.created_at AS u_created_at, u.updated_at AS u_updated_at, u.deleted_at AS u_deleted_at, u.created_by AS u_created_by, u.updated_by AS u_updated_by, u.deleted_by AS u_deleted_by";
const FROM: &str = " FROM space_members e JOIN spaces s ON s.id = e.space_id JOIN users u ON u.id = e.user_id WHERE e.deleted_at IS NULL AND s.deleted_at IS NULL AND u.deleted_at IS NULL";

pub(super) fn create(space_id: i64, input: &CreateSpaceMember) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("INSERT INTO space_members (");
    query.push("space_id, user_id, created_by");
    if input.is_active.is_some() {
        query.push(", is_active");
    }
    query.push(") SELECT ");
    {
        let mut values = query.separated(", ");
        values.push_bind(space_id);
        values.push_bind(input.user_id);
        values.push_bind(input.by);
        if let Some(value) = &input.is_active {
            values.push_bind(*value);
        }
    }
    query
        .push(" FROM spaces s JOIN users u ON u.deleted_at IS NULL WHERE s.id = ")
        .push_bind(space_id)
        .push(" AND s.deleted_at IS NULL");
    query.push(" AND u.id = ").push_bind(input.user_id);
    query.push(" RETURNING id");
    query
}

fn select(space_id: i64) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("SELECT ");
    query.push(COLUMNS).push(FROM);
    query.push(" AND e.space_id = ").push_bind(space_id);
    query
}

pub(super) fn read_by_id(space_id: i64, value: i64) -> QueryBuilder<Postgres> {
    let mut query = select(space_id);
    query.push(" AND e.id = ").push_bind(value).push(" LIMIT 1");
    query
}

pub(super) fn read_by_user_id(space_id: i64, value: i64) -> QueryBuilder<Postgres> {
    let mut query = select(space_id);
    query
        .push(" AND e.user_id = ")
        .push_bind(value)
        .push(" LIMIT 1");
    query
}

fn conditions(query: &mut QueryBuilder<Postgres>, filter: &SpaceMemberFilter) {
    if let Some(pattern) = search_pattern(&filter.search) {
        query.push(" AND (");
        query.push("u.name ILIKE ").push_bind(pattern.clone());
        query
            .push(" OR u.username ILIKE ")
            .push_bind(pattern.clone());
        query.push(" OR u.email ILIKE ").push_bind(pattern.clone());
        query.push(" OR u.phone ILIKE ").push_bind(pattern.clone());
        query.push(")");
    }
    if let Some(value) = filter.user_id {
        query.push(" AND e.user_id = ").push_bind(value);
    }
    if let Some(value) = filter.is_active {
        query.push(" AND e.is_active = ").push_bind(value);
    }
    if let Some(value) = filter.role_id {
        query.push(" AND EXISTS (SELECT 1 FROM member_role mr JOIN roles r ON r.id = mr.role_id AND r.space_id = mr.space_id WHERE mr.space_id = e.space_id AND mr.member_id = e.id AND r.deleted_at IS NULL AND mr.role_id = ").push_bind(value).push(")");
    }
}

pub(super) fn read_by_filter(
    space_id: i64,
    filter: &SpaceMemberFilter,
) -> Result<(QueryBuilder<Postgres>, QueryBuilder<Postgres>), RepositoryError> {
    let (limit, offset) = pagination(filter.page, filter.limit)?;
    let mut count = QueryBuilder::new("SELECT COUNT(*) AS total");
    count.push(FROM);
    count.push(" AND e.space_id = ").push_bind(space_id);
    let mut rows = select(space_id);
    conditions(&mut count, filter);
    conditions(&mut rows, filter);
    rows.push(" ORDER BY e.created_at DESC, e.id ASC LIMIT ")
        .push_bind(limit)
        .push(" OFFSET ")
        .push_bind(offset);
    Ok((count, rows))
}

pub(super) fn update_by_id(
    space_id: i64,
    id: i64,
    input: &UpdateSpaceMember,
) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("UPDATE space_members e SET ");
    {
        let mut sets = query.separated(", ");
        if let Some(value) = &input.is_active {
            sets.push("is_active = ").push_bind_unseparated(*value);
        }
        if let Some(value) = &input.preferences {
            sets.push("preferences = ")
                .push_bind_unseparated(value.clone());
        }
        sets.push("updated_at = CURRENT_TIMESTAMP");
        sets.push("updated_by = ").push_bind_unseparated(input.by);
    }
    query
        .push(" WHERE e.id = ")
        .push_bind(id)
        .push(" AND e.deleted_at IS NULL");
    query.push(" AND e.space_id = ").push_bind(space_id).push(
        " AND EXISTS (SELECT 1 FROM spaces s WHERE s.id = e.space_id AND s.deleted_at IS NULL)",
    );
    query.push(
        " AND EXISTS (SELECT 1 FROM users u WHERE u.id = e.user_id AND u.deleted_at IS NULL)",
    );
    query
}

pub(super) fn delete_by_id(space_id: i64, id: i64, by: Option<i64>) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "UPDATE space_members e SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ",
    );
    query
        .push_bind(by)
        .push(" WHERE e.id = ")
        .push_bind(id)
        .push(" AND e.deleted_at IS NULL");
    query.push(" AND e.space_id = ").push_bind(space_id).push(
        " AND EXISTS (SELECT 1 FROM spaces s WHERE s.id = e.space_id AND s.deleted_at IS NULL)",
    );
    query.push(
        " AND EXISTS (SELECT 1 FROM users u WHERE u.id = e.user_id AND u.deleted_at IS NULL)",
    );
    query
}

pub(super) fn lock_parents(space_id: i64, user_id: i64) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "SELECT s.id FROM spaces s JOIN users u ON u.deleted_at IS NULL WHERE s.deleted_at IS NULL AND s.id = ",
    );
    query
        .push_bind(space_id)
        .push(" AND u.id = ")
        .push_bind(user_id)
        .push(" FOR SHARE OF s, u");
    query
}
pub(super) fn lock_default(space_id: i64) -> QueryBuilder<Postgres> {
    super::super::shared::query::lock_default(space_id)
}
pub(super) fn read_default_for_share(space_id: i64) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "SELECT id FROM roles WHERE deleted_at IS NULL AND is_default = TRUE AND space_id = ",
    );
    query.push_bind(space_id).push(" FOR SHARE");
    query
}
pub(super) fn assign_default(
    space_id: i64,
    member_id: i64,
    role_id: i64,
    by: Option<i64>,
) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "INSERT INTO member_role (space_id, member_id, role_id, created_by) VALUES (",
    );
    query
        .push_bind(space_id)
        .push(", ")
        .push_bind(member_id)
        .push(", ")
        .push_bind(role_id)
        .push(", ")
        .push_bind(by)
        .push(")");
    query
}
