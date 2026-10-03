use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use crate::domain::{
    contracts::repository::{CreateRole, RoleFilter, UpdateRole},
    models::RepositoryError,
};

use super::super::shared::query::{pagination, search_pattern};

const COLUMNS: &str = "id, name, description, is_default, preferences, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by";

pub(super) fn create(input: &CreateRole) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("INSERT INTO roles (");
    query.push("name, created_by");
    if input.description.is_some() {
        query.push(", description");
    }
    if input.is_default.is_some() {
        query.push(", is_default");
    }
    query.push(") VALUES (");
    {
        let mut values = query.separated(", ");
        values.push_bind(input.name.clone());
        values.push_bind(input.by);
        if let Some(value) = &input.description {
            values.push_bind(value.clone());
        }
        if let Some(value) = &input.is_default {
            values.push_bind(*value);
        }
    }
    query.push(") RETURNING id");
    query
}

fn select() -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("SELECT ");
    query
        .push(COLUMNS)
        .push(" FROM roles WHERE deleted_at IS NULL");
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

pub(super) fn read_default() -> QueryBuilder<Postgres> {
    let mut query = select();
    query.push(" AND is_default = TRUE LIMIT 1");
    query
}

pub(super) fn lock_default() -> QueryBuilder<Postgres> {
    QueryBuilder::new("SELECT pg_advisory_xact_lock(1280262987, 1)")
}

pub(super) fn unset_default(except: Option<i64>, by: Option<i64>) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "UPDATE roles SET is_default = FALSE, updated_at = CURRENT_TIMESTAMP, updated_by = ",
    );
    query
        .push_bind(by)
        .push(" WHERE is_default = TRUE AND deleted_at IS NULL");
    if let Some(id) = except {
        query.push(" AND id <> ").push_bind(id);
    }
    query
}

fn conditions(query: &mut QueryBuilder<Postgres>, filter: &RoleFilter) {
    if let Some(pattern) = search_pattern(&filter.search) {
        query.push(" AND (");
        query.push("name ILIKE ").push_bind(pattern.clone());
        query.push(")");
    }
    if let Some(value) = filter.is_default {
        query.push(" AND is_default = ").push_bind(value);
    }
}

pub(super) fn read_by_filter(
    filter: &RoleFilter,
) -> Result<(QueryBuilder<Postgres>, QueryBuilder<Postgres>), RepositoryError> {
    let (limit, offset) = pagination(filter.page, filter.limit)?;
    let mut count =
        QueryBuilder::new("SELECT COUNT(*) AS total FROM roles WHERE deleted_at IS NULL");
    let mut rows = select();
    conditions(&mut count, filter);
    conditions(&mut rows, filter);
    rows.push(" ORDER BY created_at DESC, id ASC LIMIT ")
        .push_bind(limit)
        .push(" OFFSET ")
        .push_bind(offset);
    Ok((count, rows))
}

pub(super) fn update_by_id(id: i64, input: &UpdateRole) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("UPDATE roles SET ");
    {
        let mut sets = query.separated(", ");
        if let Some(value) = &input.name {
            sets.push("name = ").push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.description {
            sets.push("description = ")
                .push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.is_default {
            sets.push("is_default = ").push_bind_unseparated(*value);
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
    let mut query = QueryBuilder::new(
        "UPDATE roles SET is_default = FALSE, deleted_at = CURRENT_TIMESTAMP, deleted_by = ",
    );
    query
        .push_bind(by)
        .push(" WHERE id = ")
        .push_bind(id)
        .push(" AND deleted_at IS NULL");
    query
}
