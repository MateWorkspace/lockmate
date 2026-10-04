use mate_pgdt::sqlx::{self, Row, postgres::PgRow};

use crate::domain::models::Space;

pub(crate) fn decode(row: &PgRow, prefix: &str) -> Result<Space, sqlx::Error> {
    Ok(Space {
        id: row.try_get(format!("{prefix}id").as_str())?,
        slug: row.try_get(format!("{prefix}slug").as_str())?,
        name: row.try_get(format!("{prefix}name").as_str())?,
        description: row.try_get(format!("{prefix}description").as_str())?,
        is_active: row.try_get(format!("{prefix}is_active").as_str())?,
        preferences: row.try_get(format!("{prefix}preferences").as_str())?,
        audit: super::audit::full(row, prefix)?,
    })
}
