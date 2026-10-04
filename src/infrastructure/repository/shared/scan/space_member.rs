use mate_pgdt::sqlx::{self, Row, postgres::PgRow};

use crate::domain::models::SpaceMember;

pub(crate) fn decode(row: &PgRow, prefix: &str) -> Result<SpaceMember, sqlx::Error> {
    Ok(SpaceMember {
        id: row.try_get(format!("{prefix}id").as_str())?,
        space_id: row.try_get(format!("{prefix}space_id").as_str())?,
        user_id: row.try_get(format!("{prefix}user_id").as_str())?,
        is_active: row.try_get(format!("{prefix}is_active").as_str())?,
        preferences: row.try_get(format!("{prefix}preferences").as_str())?,
        audit: super::audit::full(row, prefix)?,
    })
}
