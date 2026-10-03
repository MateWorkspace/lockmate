use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use crate::domain::{contracts::repository::CreateRolePermission, models::RepositoryError};

const COLUMNS: &str = "rp.id AS rp_id, rp.role_id AS rp_role_id, rp.permission_id AS rp_permission_id, rp.created_at AS rp_created_at, rp.created_by AS rp_created_by, r.id AS r_id, r.name AS r_name, r.description AS r_description, r.is_default AS r_is_default, r.preferences AS r_preferences, r.created_at AS r_created_at, r.updated_at AS r_updated_at, r.deleted_at AS r_deleted_at, r.created_by AS r_created_by, r.updated_by AS r_updated_by, r.deleted_by AS r_deleted_by, p.id AS p_id, p.name AS p_name, p.description AS p_description, p.preferences AS p_preferences, p.created_at AS p_created_at, p.updated_at AS p_updated_at, p.deleted_at AS p_deleted_at, p.created_by AS p_created_by, p.updated_by AS p_updated_by, p.deleted_by AS p_deleted_by";

pub(super) fn create(input: &CreateRolePermission) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "INSERT INTO role_permission (role_id, permission_id, created_by) VALUES (",
    );
    query
        .push_bind(input.role_id)
        .push(", ")
        .push_bind(input.permission_id)
        .push(", ")
        .push_bind(input.by)
        .push(") RETURNING id");
    query
}

fn select() -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("SELECT ");
    query.push(COLUMNS).push(" FROM role_permission rp JOIN roles r ON rp.role_id = r.id JOIN permissions p ON rp.permission_id = p.id WHERE r.deleted_at IS NULL AND p.deleted_at IS NULL");
    query
}

pub(super) fn read_by_id(id: i64) -> QueryBuilder<Postgres> {
    let mut query = select();
    query.push(" AND rp.id = ").push_bind(id).push(" LIMIT 1");
    query
}

pub(super) fn read_by_role_id(id: i64) -> QueryBuilder<Postgres> {
    let mut query = select();
    query
        .push(" AND rp.role_id = ")
        .push_bind(id)
        .push(" ORDER BY rp.created_at DESC, rp.id ASC");
    query
}

pub(super) fn read_by_permission_id(id: i64) -> QueryBuilder<Postgres> {
    let mut query = select();
    query
        .push(" AND rp.permission_id = ")
        .push_bind(id)
        .push(" ORDER BY rp.created_at DESC, rp.id ASC");
    query
}

pub(super) fn delete_by_id(id: i64) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("DELETE FROM role_permission WHERE id = ");
    query.push_bind(id);
    query
}

pub(super) fn delete_by_role_id_or_permission_id(
    role_id: Option<i64>,
    permission_id: Option<i64>,
) -> Result<QueryBuilder<Postgres>, RepositoryError> {
    if role_id.is_none() && permission_id.is_none() {
        return Err(RepositoryError::BadArgs);
    }
    let mut query = QueryBuilder::new("DELETE FROM role_permission WHERE (");
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
