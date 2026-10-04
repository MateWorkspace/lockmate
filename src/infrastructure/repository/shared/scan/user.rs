use mate_pgdt::sqlx::{self, Row, postgres::PgRow};

use crate::domain::models::User;

pub(crate) fn decode(row: &PgRow, prefix: &str) -> Result<User, sqlx::Error> {
    Ok(User {
        id: row.try_get(format!("{prefix}id").as_str())?,
        name: row.try_get(format!("{prefix}name").as_str())?,
        bio: row.try_get(format!("{prefix}bio").as_str())?,
        username: row.try_get(format!("{prefix}username").as_str())?,
        email: row.try_get(format!("{prefix}email").as_str())?,
        phone: row.try_get(format!("{prefix}phone").as_str())?,
        password_hash: row.try_get(format!("{prefix}password_hash").as_str())?,
        is_email_verified: row.try_get(format!("{prefix}is_email_verified").as_str())?,
        is_phone_verified: row.try_get(format!("{prefix}is_phone_verified").as_str())?,
        avatar_path: row.try_get(format!("{prefix}avatar_path").as_str())?,
        preferences: row.try_get(format!("{prefix}preferences").as_str())?,
        audit: super::audit::full(row, prefix)?,
    })
}
