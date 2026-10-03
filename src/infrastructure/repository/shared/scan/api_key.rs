use mate_pgdt::sqlx::{self, Row, postgres::PgRow};

use crate::domain::models::ApiKey;

pub(crate) fn decode(row: &PgRow, prefix: &str) -> Result<ApiKey, sqlx::Error> {
    Ok(ApiKey {
        id: row.try_get(format!("{prefix}id").as_str())?,
        user_id: row.try_get(format!("{prefix}user_id").as_str())?,
        name: row.try_get(format!("{prefix}name").as_str())?,
        description: row.try_get(format!("{prefix}description").as_str())?,
        hash: row.try_get(format!("{prefix}hash").as_str())?,
        redacted: row.try_get(format!("{prefix}redacted").as_str())?,
        preferences: row.try_get(format!("{prefix}preferences").as_str())?,
        audit: super::audit::full(row, prefix)?,
    })
}
