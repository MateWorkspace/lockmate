use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use super::super::shared::query::{pagination, search_pattern};
use crate::domain::{
    contracts::repository::{CreateRole, RoleFilter, UpdateRole},
    models::RepositoryError,
};

const COLUMNS: &str = "e.id AS id, e.space_id AS space_id, e.slug AS slug, e.name AS name, e.description AS description, e.is_default AS is_default, e.preferences AS preferences, e.created_at AS created_at, e.updated_at AS updated_at, e.deleted_at AS deleted_at, e.created_by AS created_by, e.updated_by AS updated_by, e.deleted_by AS deleted_by";
const FROM: &str = " FROM roles e JOIN spaces s ON s.id = e.space_id WHERE e.deleted_at IS NULL AND s.deleted_at IS NULL";

pub(super) fn create(space_id: i64, input: &CreateRole) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("INSERT INTO roles (");
    query.push("space_id, slug, name, created_by");
    if input.description.is_some() {
        query.push(", description");
    }
    if input.is_default.is_some() {
        query.push(", is_default");
    }
    query.push(") SELECT ");
    {
        let mut values = query.separated(", ");
        values.push_bind(space_id);
        values.push_bind(input.slug.clone());
        values.push_bind(input.name.clone());
        values.push_bind(input.by);
        if let Some(value) = &input.description {
            values.push_bind(value.clone());
        }
        if let Some(value) = &input.is_default {
            values.push_bind(*value);
        }
    }
    query
        .push(" FROM spaces s WHERE s.id = ")
        .push_bind(space_id)
        .push(" AND s.deleted_at IS NULL");
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

pub(super) fn read_by_slug(space_id: i64, value: &str) -> QueryBuilder<Postgres> {
    let mut query = select(space_id);
    query
        .push(" AND e.slug = ")
        .push_bind(value)
        .push(" LIMIT 1");
    query
}

fn conditions(query: &mut QueryBuilder<Postgres>, filter: &RoleFilter) {
    if let Some(pattern) = search_pattern(&filter.search) {
        query.push(" AND (");
        query.push("e.slug ILIKE ").push_bind(pattern.clone());
        query.push(" OR e.name ILIKE ").push_bind(pattern.clone());
        query.push(")");
    }
    if let Some(value) = filter.is_default {
        query.push(" AND e.is_default = ").push_bind(value);
    }
}

pub(super) fn read_by_filter(
    space_id: i64,
    filter: &RoleFilter,
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

pub(super) fn update_by_id(space_id: i64, id: i64, input: &UpdateRole) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("UPDATE roles e SET ");
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
        .push(" WHERE e.id = ")
        .push_bind(id)
        .push(" AND e.deleted_at IS NULL");
    query.push(" AND e.space_id = ").push_bind(space_id).push(
        " AND EXISTS (SELECT 1 FROM spaces s WHERE s.id = e.space_id AND s.deleted_at IS NULL)",
    );
    query
}

pub(super) fn delete_by_id(space_id: i64, id: i64, by: Option<i64>) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "UPDATE roles e SET is_default = FALSE, deleted_at = CURRENT_TIMESTAMP, deleted_by = ",
    );
    query
        .push_bind(by)
        .push(" WHERE e.id = ")
        .push_bind(id)
        .push(" AND e.deleted_at IS NULL");
    query.push(" AND e.space_id = ").push_bind(space_id).push(
        " AND EXISTS (SELECT 1 FROM spaces s WHERE s.id = e.space_id AND s.deleted_at IS NULL)",
    );
    query
}

pub(super) fn read_default(space_id: i64) -> QueryBuilder<Postgres> {
    let mut query = select(space_id);
    query.push(" AND e.is_default = TRUE LIMIT 1");
    query
}
pub(super) fn lock_default(space_id: i64) -> QueryBuilder<Postgres> {
    super::super::shared::query::lock_default(space_id)
}
pub(super) fn unset_default(
    space_id: i64,
    except: Option<i64>,
    by: Option<i64>,
) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "UPDATE roles SET is_default = FALSE, updated_at = CURRENT_TIMESTAMP, updated_by = ",
    );
    query
        .push_bind(by)
        .push(" WHERE space_id = ")
        .push_bind(space_id)
        .push(" AND is_default = TRUE AND deleted_at IS NULL");
    if let Some(id) = except {
        query.push(" AND id <> ").push_bind(id);
    }
    query
}
