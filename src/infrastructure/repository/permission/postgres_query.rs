use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use crate::domain::{
    contracts::repository::{CreatePermission, PermissionFilter, UpdatePermission},
    models::RepositoryError,
};

use super::super::shared::query::{pagination, search_pattern};

const COLUMNS: &str = "id, name, description, preferences, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by";

pub(super) fn create(input: &CreatePermission) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("INSERT INTO permissions (");
    query.push("name, created_by");
    if input.description.is_some() {
        query.push(", description");
    }
    query.push(") VALUES (");
    {
        let mut values = query.separated(", ");
        values.push_bind(input.name.clone());
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
        .push(" FROM permissions WHERE deleted_at IS NULL");
    query
}

pub(super) fn read_by_id(value: i64) -> QueryBuilder<Postgres> {
    let mut query = select();
    query.push(" AND id = ").push_bind(value).push(" LIMIT 1");
    query
}

pub(super) fn read_by_name(value: &str) -> QueryBuilder<Postgres> {
    let mut query = select();
    query.push(" AND name = ").push_bind(value).push(" LIMIT 1");
    query
}

fn conditions(query: &mut QueryBuilder<Postgres>, filter: &PermissionFilter) {
    if let Some(pattern) = search_pattern(&filter.search) {
        query.push(" AND (");
        query.push("name ILIKE ").push_bind(pattern.clone());
        query.push(")");
    }
}

pub(super) fn read_by_filter(
    filter: &PermissionFilter,
) -> Result<(QueryBuilder<Postgres>, QueryBuilder<Postgres>), RepositoryError> {
    let (limit, offset) = pagination(filter.page, filter.limit)?;
    let mut count =
        QueryBuilder::new("SELECT COUNT(*) AS total FROM permissions WHERE deleted_at IS NULL");
    let mut rows = select();
    conditions(&mut count, filter);
    conditions(&mut rows, filter);
    rows.push(" ORDER BY created_at DESC, id ASC LIMIT ")
        .push_bind(limit)
        .push(" OFFSET ")
        .push_bind(offset);
    Ok((count, rows))
}

pub(super) fn update_by_id(id: i64, input: &UpdatePermission) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("UPDATE permissions SET ");
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
        QueryBuilder::new("UPDATE permissions SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ");
    query
        .push_bind(by)
        .push(" WHERE id = ")
        .push_bind(id)
        .push(" AND deleted_at IS NULL");
    query
}
