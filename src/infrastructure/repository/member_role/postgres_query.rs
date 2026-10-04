use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use crate::domain::{contracts::repository::CreateMemberRole, models::RepositoryError};

const COLUMNS: &str = "e.id AS mr_id, e.space_id AS mr_space_id, e.member_id AS mr_member_id, e.role_id AS mr_role_id, e.created_at AS mr_created_at, e.created_by AS mr_created_by, m.id AS m_id, m.space_id AS m_space_id, m.user_id AS m_user_id, m.is_active AS m_is_active, m.preferences AS m_preferences, m.created_at AS m_created_at, m.updated_at AS m_updated_at, m.deleted_at AS m_deleted_at, m.created_by AS m_created_by, m.updated_by AS m_updated_by, m.deleted_by AS m_deleted_by, r.id AS r_id, r.space_id AS r_space_id, r.slug AS r_slug, r.name AS r_name, r.description AS r_description, r.is_default AS r_is_default, r.preferences AS r_preferences, r.created_at AS r_created_at, r.updated_at AS r_updated_at, r.deleted_at AS r_deleted_at, r.created_by AS r_created_by, r.updated_by AS r_updated_by, r.deleted_by AS r_deleted_by";
const FROM: &str = " FROM member_role e JOIN space_members m ON m.id = e.member_id AND m.space_id = e.space_id JOIN roles r ON r.id = e.role_id AND r.space_id = e.space_id JOIN spaces s ON s.id = e.space_id JOIN users u ON u.id = m.user_id WHERE m.deleted_at IS NULL AND r.deleted_at IS NULL AND s.deleted_at IS NULL AND u.deleted_at IS NULL";

pub(super) fn create(space_id: i64, input: &CreateMemberRole) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("INSERT INTO member_role (");
    query.push("space_id, member_id, role_id, created_by");
    query.push(") SELECT ");
    {
        let mut values = query.separated(", ");
        values.push_bind(space_id);
        values.push_bind(input.member_id);
        values.push_bind(input.role_id);
        values.push_bind(input.by);
    }
    query.push(" FROM spaces s JOIN space_members m ON m.space_id = s.id JOIN users u ON u.id = m.user_id JOIN roles r ON r.space_id = s.id WHERE s.id = ").push_bind(space_id).push(" AND s.deleted_at IS NULL");
    query
        .push(" AND m.id = ")
        .push_bind(input.member_id)
        .push(" AND r.id = ")
        .push_bind(input.role_id)
        .push(" AND m.deleted_at IS NULL AND u.deleted_at IS NULL AND r.deleted_at IS NULL");
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

pub(super) fn read_by_member_id(space_id: i64, id: i64) -> QueryBuilder<Postgres> {
    let mut query = select(space_id);
    query
        .push(" AND e.member_id = ")
        .push_bind(id)
        .push(" ORDER BY e.created_at DESC, e.id ASC");
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

pub(super) fn read_by_member_id_and_role_id(
    space_id: i64,
    member_id: i64,
    role_id: i64,
) -> QueryBuilder<Postgres> {
    let mut query = select(space_id);
    query
        .push(" AND e.member_id = ")
        .push_bind(member_id)
        .push(" AND e.role_id = ")
        .push_bind(role_id)
        .push(" LIMIT 1");
    query
}

pub(super) fn delete_by_id(space_id: i64, id: i64) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("DELETE FROM member_role WHERE space_id = ");
    query.push_bind(space_id).push(" AND id = ").push_bind(id);
    query
}

pub(super) fn delete_by_member_id_or_role_id(
    space_id: i64,
    member_id: Option<i64>,
    role_id: Option<i64>,
) -> Result<QueryBuilder<Postgres>, RepositoryError> {
    if member_id.is_none() && role_id.is_none() {
        return Err(RepositoryError::BadArgs);
    }
    let mut query = QueryBuilder::new("DELETE FROM member_role WHERE space_id = ");
    query.push_bind(space_id).push(" AND (");
    if let Some(id) = member_id {
        query.push("member_id = ").push_bind(id);
    }
    if let Some(id) = role_id {
        if member_id.is_some() {
            query.push(" OR ");
        }
        query.push("role_id = ").push_bind(id);
    }
    query.push(")");
    Ok(query)
}

pub(super) fn access_state(space_id: i64, member_id: i64) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "SELECT s.is_active AS space_active, m.is_active AS member_active FROM space_members m JOIN spaces s ON s.id = m.space_id JOIN users u ON u.id = m.user_id WHERE m.deleted_at IS NULL AND s.deleted_at IS NULL AND u.deleted_at IS NULL AND m.space_id = ",
    );
    query
        .push_bind(space_id)
        .push(" AND m.id = ")
        .push_bind(member_id);
    query
}

pub(super) fn read_effective_roles_by_member_id(
    space_id: i64,
    member_id: i64,
) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "SELECT DISTINCT r.id AS id, r.space_id AS space_id, r.slug AS slug, r.name AS name, r.description AS description, r.is_default AS is_default, r.preferences AS preferences, r.created_at AS created_at, r.updated_at AS updated_at, r.deleted_at AS deleted_at, r.created_by AS created_by, r.updated_by AS updated_by, r.deleted_by AS deleted_by FROM member_role e JOIN space_members m ON m.id = e.member_id AND m.space_id = e.space_id JOIN roles r ON r.id = e.role_id AND r.space_id = e.space_id JOIN spaces s ON s.id = e.space_id JOIN users u ON u.id = m.user_id WHERE m.deleted_at IS NULL AND r.deleted_at IS NULL AND s.deleted_at IS NULL AND u.deleted_at IS NULL",
    );
    query
        .push(" AND e.space_id = ")
        .push_bind(space_id)
        .push(" AND e.member_id = ")
        .push_bind(member_id)
        .push(" AND s.is_active = TRUE AND m.is_active = TRUE ORDER BY r.slug ASC, r.id ASC");
    query
}

pub(super) fn read_effective_permissions_by_member_id(
    space_id: i64,
    member_id: i64,
) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "SELECT DISTINCT p.id AS id, p.space_id AS space_id, p.slug AS slug, p.name AS name, p.description AS description, p.preferences AS preferences, p.created_at AS created_at, p.updated_at AS updated_at, p.deleted_at AS deleted_at, p.created_by AS created_by, p.updated_by AS updated_by, p.deleted_by AS deleted_by FROM member_role e JOIN space_members m ON m.id = e.member_id AND m.space_id = e.space_id JOIN roles r ON r.id = e.role_id AND r.space_id = e.space_id JOIN spaces s ON s.id = e.space_id JOIN users u ON u.id = m.user_id JOIN role_permission rp ON rp.role_id = r.id AND rp.space_id = e.space_id JOIN permissions p ON p.id = rp.permission_id AND p.space_id = e.space_id WHERE m.deleted_at IS NULL AND r.deleted_at IS NULL AND s.deleted_at IS NULL AND u.deleted_at IS NULL",
    );
    query.push(" AND e.space_id = ").push_bind(space_id).push(" AND e.member_id = ").push_bind(member_id).push(" AND s.is_active = TRUE AND m.is_active = TRUE AND p.deleted_at IS NULL ORDER BY p.slug ASC, p.id ASC");
    query
}
