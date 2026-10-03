use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use crate::domain::{
    contracts::repository::{ApiKeyFilter, CreateApiKey, UpdateApiKey},
    models::RepositoryError,
};

use super::super::shared::query::{pagination, search_pattern};

const COLUMNS: &str = "id, user_id, name, description, hash, redacted, preferences, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by";

pub(super) fn create(input: &CreateApiKey) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("INSERT INTO api_keys (");
    query.push("user_id, name, hash, redacted, created_by");
    if input.description.is_some() {
        query.push(", description");
    }
    query.push(") VALUES (");
    {
        let mut values = query.separated(", ");
        values.push_bind(input.user_id);
        values.push_bind(input.name.clone());
        values.push_bind(input.hash.clone());
        values.push_bind(input.redacted.clone());
        values.push_bind(input.by);
        if let Some(value) = &input.description {
            values.push_bind(value.clone());
        }
    }
    query.push(") RETURNING id");
    query
}

fn select() -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("SELECT ");
    query
        .push(COLUMNS)
        .push(" FROM api_keys WHERE deleted_at IS NULL");
    query
}

pub(super) fn read_by_id(value: i64) -> QueryBuilder<Postgres> {
    let mut query = select();
    query.push(" AND id = ").push_bind(value).push(" LIMIT 1");
    query
}

pub(super) fn read_by_hash(value: &str) -> QueryBuilder<Postgres> {
    let mut query = select();
    query.push(" AND hash = ").push_bind(value).push(" LIMIT 1");
    query
}

fn conditions(query: &mut QueryBuilder<Postgres>, filter: &ApiKeyFilter) {
    if let Some(pattern) = search_pattern(&filter.search) {
        query.push(" AND (");
        query.push("name ILIKE ").push_bind(pattern.clone());
        query.push(" OR redacted ILIKE ").push_bind(pattern.clone());
        query.push(")");
    }
    if let Some(value) = filter.user_id {
        query.push(" AND user_id = ").push_bind(value);
    }
}

pub(super) fn read_by_filter(
    filter: &ApiKeyFilter,
) -> Result<(QueryBuilder<Postgres>, QueryBuilder<Postgres>), RepositoryError> {
    let (limit, offset) = pagination(filter.page, filter.limit)?;
    let mut count =
        QueryBuilder::new("SELECT COUNT(*) AS total FROM api_keys WHERE deleted_at IS NULL");
    let mut rows = select();
    conditions(&mut count, filter);
    conditions(&mut rows, filter);
    rows.push(" ORDER BY created_at DESC, id ASC LIMIT ")
        .push_bind(limit)
        .push(" OFFSET ")
        .push_bind(offset);
    Ok((count, rows))
}

pub(super) fn update_by_id(id: i64, input: &UpdateApiKey) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("UPDATE api_keys SET ");
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
        .push(" WHERE id = ")
        .push_bind(id)
        .push(" AND deleted_at IS NULL");
    query
}

pub(super) fn delete_by_id(id: i64, by: Option<i64>) -> QueryBuilder<Postgres> {
    let mut query =
        QueryBuilder::new("UPDATE api_keys SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ");
    query
        .push_bind(by)
        .push(" WHERE id = ")
        .push_bind(id)
        .push(" AND deleted_at IS NULL");
    query
}
