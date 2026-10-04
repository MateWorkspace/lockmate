use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use super::super::shared::query::{pagination, search_pattern};
use crate::domain::{
    contracts::repository::{CreateSpace, SpaceFilter, UpdateSpace},
    models::RepositoryError,
};

const COLUMNS: &str = "e.id AS id, e.slug AS slug, e.name AS name, e.description AS description, e.is_active AS is_active, e.preferences AS preferences, e.created_at AS created_at, e.updated_at AS updated_at, e.deleted_at AS deleted_at, e.created_by AS created_by, e.updated_by AS updated_by, e.deleted_by AS deleted_by";
const FROM: &str = " FROM spaces e WHERE e.deleted_at IS NULL";

pub(super) fn create(input: &CreateSpace) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("INSERT INTO spaces (");
    query.push("slug, name, created_by");
    if input.description.is_some() {
        query.push(", description");
    }
    if input.is_active.is_some() {
        query.push(", is_active");
    }
    query.push(") SELECT ");
    {
        let mut values = query.separated(", ");
        values.push_bind(input.slug.clone());
        values.push_bind(input.name.clone());
        values.push_bind(input.by);
        if let Some(value) = &input.description {
            values.push_bind(value.clone());
        }
        if let Some(value) = &input.is_active {
            values.push_bind(*value);
        }
    }
    query.push(" RETURNING id");
    query
}

fn select() -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("SELECT ");
    query.push(COLUMNS).push(FROM);
    query
}

pub(super) fn read_by_id(value: i64) -> QueryBuilder<Postgres> {
    let mut query = select();
    query.push(" AND e.id = ").push_bind(value).push(" LIMIT 1");
    query
}

pub(super) fn read_by_slug(value: &str) -> QueryBuilder<Postgres> {
    let mut query = select();
    query
        .push(" AND e.slug = ")
        .push_bind(value)
        .push(" LIMIT 1");
    query
}

fn conditions(query: &mut QueryBuilder<Postgres>, filter: &SpaceFilter) {
    if let Some(pattern) = search_pattern(&filter.search) {
        query.push(" AND (");
        query.push("e.slug ILIKE ").push_bind(pattern.clone());
        query.push(" OR e.name ILIKE ").push_bind(pattern.clone());
        query.push(")");
    }
    if let Some(value) = filter.is_active {
        query.push(" AND e.is_active = ").push_bind(value);
    }
}

pub(super) fn read_by_filter(
    filter: &SpaceFilter,
) -> Result<(QueryBuilder<Postgres>, QueryBuilder<Postgres>), RepositoryError> {
    let (limit, offset) = pagination(filter.page, filter.limit)?;
    let mut count = QueryBuilder::new("SELECT COUNT(*) AS total");
    count.push(FROM);
    let mut rows = select();
    conditions(&mut count, filter);
    conditions(&mut rows, filter);
    rows.push(" ORDER BY e.created_at DESC, e.id ASC LIMIT ")
        .push_bind(limit)
        .push(" OFFSET ")
        .push_bind(offset);
    Ok((count, rows))
}

pub(super) fn update_by_id(id: i64, input: &UpdateSpace) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("UPDATE spaces e SET ");
    {
        let mut sets = query.separated(", ");
        if let Some(value) = &input.name {
            sets.push("name = ").push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.description {
            sets.push("description = ")
                .push_bind_unseparated(value.clone());
        }
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
    query
}

pub(super) fn delete_by_id(id: i64, by: Option<i64>) -> QueryBuilder<Postgres> {
    let mut query =
        QueryBuilder::new("UPDATE spaces e SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ");
    query
        .push_bind(by)
        .push(" WHERE e.id = ")
        .push_bind(id)
        .push(" AND e.deleted_at IS NULL");
    query
}
