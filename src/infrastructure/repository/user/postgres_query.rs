use mate_pgdt::sqlx::{Postgres, QueryBuilder};

use crate::domain::{
    contracts::repository::{CreateUser, UpdateUser, UserFilter},
    models::RepositoryError,
};

use super::super::shared::query::{pagination, search_pattern};

const COLUMNS: &str = "id, role_id, name, bio, username, email, phone, password_hash, is_email_verified, is_phone_verified, avatar_path, preferences, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by";

pub(super) fn create(input: &CreateUser) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("INSERT INTO users (");
    query.push("role_id, name, username, password_hash, created_by");
    if input.bio.is_some() {
        query.push(", bio");
    }
    if input.email.is_some() {
        query.push(", email");
    }
    if input.phone.is_some() {
        query.push(", phone");
    }
    if input.is_email_verified.is_some() {
        query.push(", is_email_verified");
    }
    if input.is_phone_verified.is_some() {
        query.push(", is_phone_verified");
    }
    if input.avatar_path.is_some() {
        query.push(", avatar_path");
    }
    query.push(") VALUES (");
    {
        let mut values = query.separated(", ");
        values.push_bind(input.role_id);
        values.push_bind(input.name.clone());
        values.push_bind(input.username.clone());
        values.push_bind(input.password_hash.clone());
        values.push_bind(input.by);
        if let Some(value) = &input.bio {
            values.push_bind(value.clone());
        }
        if let Some(value) = &input.email {
            values.push_bind(value.clone());
        }
        if let Some(value) = &input.phone {
            values.push_bind(value.clone());
        }
        if let Some(value) = &input.is_email_verified {
            values.push_bind(*value);
        }
        if let Some(value) = &input.is_phone_verified {
            values.push_bind(*value);
        }
        if let Some(value) = &input.avatar_path {
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
        .push(" FROM users WHERE deleted_at IS NULL");
    query
}

pub(super) fn read_by_id(value: i64) -> QueryBuilder<Postgres> {
    let mut query = select();
    query.push(" AND id = ").push_bind(value).push(" LIMIT 1");
    query
}

pub(super) fn read_by_username(value: &str) -> QueryBuilder<Postgres> {
    let mut query = select();
    query
        .push(" AND username = ")
        .push_bind(value)
        .push(" LIMIT 1");
    query
}

pub(super) fn read_by_email(value: &str) -> QueryBuilder<Postgres> {
    let mut query = select();
    query
        .push(" AND email = ")
        .push_bind(value)
        .push(" LIMIT 1");
    query
}

pub(super) fn read_by_phone(value: &str) -> QueryBuilder<Postgres> {
    let mut query = select();
    query
        .push(" AND phone = ")
        .push_bind(value)
        .push(" LIMIT 1");
    query
}

fn conditions(query: &mut QueryBuilder<Postgres>, filter: &UserFilter) {
    if let Some(pattern) = search_pattern(&filter.search) {
        query.push(" AND (");
        query.push("name ILIKE ").push_bind(pattern.clone());
        query.push(" OR username ILIKE ").push_bind(pattern.clone());
        query.push(" OR email ILIKE ").push_bind(pattern.clone());
        query.push(" OR phone ILIKE ").push_bind(pattern.clone());
        query.push(")");
    }
    if let Some(value) = filter.role_id {
        query.push(" AND role_id = ").push_bind(value);
    }
    if let Some(value) = filter.is_email_verified {
        query.push(" AND is_email_verified = ").push_bind(value);
    }
    if let Some(value) = filter.is_phone_verified {
        query.push(" AND is_phone_verified = ").push_bind(value);
    }
}

pub(super) fn read_by_filter(
    filter: &UserFilter,
) -> Result<(QueryBuilder<Postgres>, QueryBuilder<Postgres>), RepositoryError> {
    let (limit, offset) = pagination(filter.page, filter.limit)?;
    let mut count =
        QueryBuilder::new("SELECT COUNT(*) AS total FROM users WHERE deleted_at IS NULL");
    let mut rows = select();
    conditions(&mut count, filter);
    conditions(&mut rows, filter);
    rows.push(" ORDER BY created_at DESC, id ASC LIMIT ")
        .push_bind(limit)
        .push(" OFFSET ")
        .push_bind(offset);
    Ok((count, rows))
}

pub(super) fn update_by_id(id: i64, input: &UpdateUser) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new("UPDATE users SET ");
    {
        let mut sets = query.separated(", ");
        if let Some(value) = &input.role_id {
            sets.push("role_id = ").push_bind_unseparated(*value);
        }
        if let Some(value) = &input.name {
            sets.push("name = ").push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.bio {
            sets.push("bio = ").push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.username {
            sets.push("username = ")
                .push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.email {
            sets.push("email = ").push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.phone {
            sets.push("phone = ").push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.password_hash {
            sets.push("password_hash = ")
                .push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.is_email_verified {
            sets.push("is_email_verified = ")
                .push_bind_unseparated(*value);
        }
        if let Some(value) = &input.is_phone_verified {
            sets.push("is_phone_verified = ")
                .push_bind_unseparated(*value);
        }
        if let Some(value) = &input.avatar_path {
            sets.push("avatar_path = ")
                .push_bind_unseparated(value.clone());
        }
        if let Some(value) = &input.preferences {
            sets.push("preferences = ")
                .push_bind_unseparated(value.clone());
        }
        if input.is_email_verified.is_none() {
            match &input.email {
                Some(None) => {
                    sets.push("is_email_verified = FALSE");
                }
                Some(Some(value)) => {
                    sets.push("is_email_verified = CASE WHEN email IS DISTINCT FROM ")
                        .push_bind_unseparated(value.clone())
                        .push_unseparated(" THEN FALSE ELSE is_email_verified END");
                }
                None => {}
            }
        }
        if input.is_phone_verified.is_none() {
            match &input.phone {
                Some(None) => {
                    sets.push("is_phone_verified = FALSE");
                }
                Some(Some(value)) => {
                    sets.push("is_phone_verified = CASE WHEN phone IS DISTINCT FROM ")
                        .push_bind_unseparated(value.clone())
                        .push_unseparated(" THEN FALSE ELSE is_phone_verified END");
                }
                None => {}
            }
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
        QueryBuilder::new("UPDATE users SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ");
    query
        .push_bind(by)
        .push(" WHERE id = ")
        .push_bind(id)
        .push(" AND deleted_at IS NULL");
    query
}
