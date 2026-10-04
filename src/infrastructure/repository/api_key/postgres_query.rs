use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use super::super::shared::query::{pagination, search_pattern};
use crate::domain::{
    contracts::repository::{ApiKeyFilter, CreateApiKey, UpdateApiKey},
    models::RepositoryError,
};

const COLUMNS: &str = "e.id AS id, e.space_id AS space_id, e.member_id AS member_id, e.name AS name, e.description AS description, e.hash AS hash, e.redacted AS redacted, e.preferences AS preferences, e.created_at AS created_at, e.updated_at AS updated_at, e.deleted_at AS deleted_at, e.created_by AS created_by, e.updated_by AS updated_by, e.deleted_by AS deleted_by";
const FROM: &str = " FROM api_keys e JOIN space_members m ON m.id = e.member_id AND m.space_id = e.space_id JOIN spaces s ON s.id = e.space_id JOIN users u ON u.id = m.user_id WHERE e.deleted_at IS NULL AND m.deleted_at IS NULL AND s.deleted_at IS NULL AND u.deleted_at IS NULL";

pub(super) fn create(space_id: i64, input: &CreateApiKey) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("INSERT INTO api_keys (");
    query.push("space_id, member_id, name, hash, redacted, created_by");
    if input.description.is_some() {
        query.push(", description");
    }
    query.push(") SELECT ");
    {
        let mut values = query.separated(", ");
        values.push_bind(space_id);
        values.push_bind(input.member_id);
        values.push_bind(input.name.clone());
        values.push_bind(input.hash.clone());
        values.push_bind(input.redacted.clone());
        values.push_bind(input.by);
        if let Some(value) = &input.description {
            values.push_bind(value.clone());
        }
    }
    query.push(" FROM spaces s JOIN space_members m ON m.space_id = s.id JOIN users u ON u.id = m.user_id WHERE s.id = ").push_bind(space_id).push(" AND s.deleted_at IS NULL");
    query
        .push(" AND m.id = ")
        .push_bind(input.member_id)
        .push(" AND m.deleted_at IS NULL AND u.deleted_at IS NULL");
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

pub(super) fn read_by_hash(space_id: i64, value: &str) -> QueryBuilder<Postgres> {
    let mut query = select(space_id);
    query
        .push(" AND e.hash = ")
        .push_bind(value)
        .push(" LIMIT 1");
    query
}

fn conditions(query: &mut QueryBuilder<Postgres>, filter: &ApiKeyFilter) {
    if let Some(pattern) = search_pattern(&filter.search) {
        query.push(" AND (");
        query.push("e.name ILIKE ").push_bind(pattern.clone());
        query
            .push(" OR e.redacted ILIKE ")
            .push_bind(pattern.clone());
        query.push(")");
    }
    if let Some(value) = filter.member_id {
        query.push(" AND e.member_id = ").push_bind(value);
    }
    if let Some(value) = filter.user_id {
        query.push(" AND m.user_id = ").push_bind(value);
    }
}

pub(super) fn read_by_filter(
    space_id: i64,
    filter: &ApiKeyFilter,
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

pub(super) fn update_by_id(space_id: i64, id: i64, input: &UpdateApiKey) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("UPDATE api_keys e SET ");
    {
        let mut sets = query.separated(", ");
        if let Some(value) = &input.name {
            sets.push("name = ").push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.description {
            sets.push("description = ")
                .push_bind_unseparated(value.clone());
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
    query.push(" AND EXISTS (SELECT 1 FROM space_members m JOIN users u ON u.id = m.user_id WHERE m.id = e.member_id AND m.space_id = e.space_id AND m.deleted_at IS NULL AND u.deleted_at IS NULL)");
    query
}

pub(super) fn delete_by_id(space_id: i64, id: i64, by: Option<i64>) -> QueryBuilder<Postgres> {
    let mut query =
        QueryBuilder::new("UPDATE api_keys e SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ");
    query
        .push_bind(by)
        .push(" WHERE e.id = ")
        .push_bind(id)
        .push(" AND e.deleted_at IS NULL");
    query.push(" AND e.space_id = ").push_bind(space_id).push(
        " AND EXISTS (SELECT 1 FROM spaces s WHERE s.id = e.space_id AND s.deleted_at IS NULL)",
    );
    query.push(" AND EXISTS (SELECT 1 FROM space_members m JOIN users u ON u.id = m.user_id WHERE m.id = e.member_id AND m.space_id = e.space_id AND m.deleted_at IS NULL AND u.deleted_at IS NULL)");
    query
}

pub(super) fn read_active_by_hash(space_id: i64, hash: &str) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "SELECT e.id AS k_id, e.space_id AS k_space_id, e.member_id AS k_member_id, e.name AS k_name, e.description AS k_description, e.hash AS k_hash, e.redacted AS k_redacted, e.preferences AS k_preferences, e.created_at AS k_created_at, e.updated_at AS k_updated_at, e.deleted_at AS k_deleted_at, e.created_by AS k_created_by, e.updated_by AS k_updated_by, e.deleted_by AS k_deleted_by, m.id AS m_id, m.space_id AS m_space_id, m.user_id AS m_user_id, m.is_active AS m_is_active, m.preferences AS m_preferences, m.created_at AS m_created_at, m.updated_at AS m_updated_at, m.deleted_at AS m_deleted_at, m.created_by AS m_created_by, m.updated_by AS m_updated_by, m.deleted_by AS m_deleted_by, u.id AS u_id, u.name AS u_name, u.bio AS u_bio, u.username AS u_username, u.email AS u_email, u.phone AS u_phone, u.password_hash AS u_password_hash, u.is_email_verified AS u_is_email_verified, u.is_phone_verified AS u_is_phone_verified, u.avatar_path AS u_avatar_path, u.preferences AS u_preferences, u.created_at AS u_created_at, u.updated_at AS u_updated_at, u.deleted_at AS u_deleted_at, u.created_by AS u_created_by, u.updated_by AS u_updated_by, u.deleted_by AS u_deleted_by, s.id AS s_id, s.slug AS s_slug, s.name AS s_name, s.description AS s_description, s.is_active AS s_is_active, s.preferences AS s_preferences, s.created_at AS s_created_at, s.updated_at AS s_updated_at, s.deleted_at AS s_deleted_at, s.created_by AS s_created_by, s.updated_by AS s_updated_by, s.deleted_by AS s_deleted_by FROM api_keys e JOIN space_members m ON m.id = e.member_id AND m.space_id = e.space_id JOIN spaces s ON s.id = e.space_id JOIN users u ON u.id = m.user_id WHERE e.deleted_at IS NULL AND m.deleted_at IS NULL AND s.deleted_at IS NULL AND u.deleted_at IS NULL",
    );
    query
        .push(" AND e.space_id = ")
        .push_bind(space_id)
        .push(" AND e.hash = ")
        .push_bind(hash)
        .push(" AND s.is_active = TRUE AND m.is_active = TRUE LIMIT 1");
    query
}
