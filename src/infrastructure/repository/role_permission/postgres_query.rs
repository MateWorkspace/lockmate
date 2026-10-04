use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use crate::domain::{contracts::repository::CreateRolePermission, models::RepositoryError};

const COLUMNS: &str = "e.id AS rp_id, e.space_id AS rp_space_id, e.role_id AS rp_role_id, e.permission_id AS rp_permission_id, e.created_at AS rp_created_at, e.created_by AS rp_created_by, r.id AS r_id, r.space_id AS r_space_id, r.slug AS r_slug, r.name AS r_name, r.description AS r_description, r.is_default AS r_is_default, r.preferences AS r_preferences, r.created_at AS r_created_at, r.updated_at AS r_updated_at, r.deleted_at AS r_deleted_at, r.created_by AS r_created_by, r.updated_by AS r_updated_by, r.deleted_by AS r_deleted_by, p.id AS p_id, p.space_id AS p_space_id, p.slug AS p_slug, p.name AS p_name, p.description AS p_description, p.preferences AS p_preferences, p.created_at AS p_created_at, p.updated_at AS p_updated_at, p.deleted_at AS p_deleted_at, p.created_by AS p_created_by, p.updated_by AS p_updated_by, p.deleted_by AS p_deleted_by";
const FROM: &str = " FROM role_permission e JOIN roles r ON r.id = e.role_id AND r.space_id = e.space_id JOIN permissions p ON p.id = e.permission_id AND p.space_id = e.space_id JOIN spaces s ON s.id = e.space_id WHERE r.deleted_at IS NULL AND p.deleted_at IS NULL AND s.deleted_at IS NULL";

pub(super) fn create(space_id: i64, input: &CreateRolePermission) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("INSERT INTO role_permission (");
    query.push("space_id, role_id, permission_id, created_by");
    query.push(") SELECT ");
    {
        let mut values = query.separated(", ");
        values.push_bind(space_id);
        values.push_bind(input.role_id);
        values.push_bind(input.permission_id);
        values.push_bind(input.by);
    }
    query.push(" FROM spaces s JOIN roles r ON r.space_id = s.id JOIN permissions p ON p.space_id = s.id WHERE s.id = ").push_bind(space_id).push(" AND s.deleted_at IS NULL");
    query
        .push(" AND r.id = ")
        .push_bind(input.role_id)
        .push(" AND p.id = ")
        .push_bind(input.permission_id)
        .push(" AND r.deleted_at IS NULL AND p.deleted_at IS NULL");
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

pub(super) fn read_by_role_id(space_id: i64, id: i64) -> QueryBuilder<Postgres> {
    let mut query = select(space_id);
    query
        .push(" AND e.role_id = ")
        .push_bind(id)
        .push(" ORDER BY e.created_at DESC, e.id ASC");
    query
}

pub(super) fn read_by_permission_id(space_id: i64, id: i64) -> QueryBuilder<Postgres> {
    let mut query = select(space_id);
    query
        .push(" AND e.permission_id = ")
        .push_bind(id)
        .push(" ORDER BY e.created_at DESC, e.id ASC");
    query
}

pub(super) fn read_by_role_id_and_permission_id(
    space_id: i64,
    role_id: i64,
    permission_id: i64,
) -> QueryBuilder<Postgres> {
    let mut query = select(space_id);
    query
        .push(" AND e.role_id = ")
        .push_bind(role_id)
        .push(" AND e.permission_id = ")
        .push_bind(permission_id)
        .push(" LIMIT 1");
    query
}

pub(super) fn delete_by_id(space_id: i64, id: i64) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("DELETE FROM role_permission WHERE space_id = ");
    query.push_bind(space_id).push(" AND id = ").push_bind(id);
    query
}

pub(super) fn delete_by_role_id_or_permission_id(
    space_id: i64,
    role_id: Option<i64>,
    permission_id: Option<i64>,
) -> Result<QueryBuilder<Postgres>, RepositoryError> {
    if role_id.is_none() && permission_id.is_none() {
        return Err(RepositoryError::BadArgs);
    }
    let mut query = QueryBuilder::new("DELETE FROM role_permission WHERE space_id = ");
    query.push_bind(space_id).push(" AND (");
    if let Some(id) = role_id {
        query.push("role_id = ").push_bind(id);
    }
    if let Some(id) = permission_id {
        if role_id.is_some() {
            query.push(" OR ");
        }
        query.push("permission_id = ").push_bind(id);
    }
    query.push(")");
    Ok(query)
}
