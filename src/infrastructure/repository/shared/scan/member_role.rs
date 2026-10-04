use mate_pgdt::sqlx::{self, Row, postgres::PgRow};

use crate::domain::models::MemberRole;

pub(crate) fn decode(row: &PgRow, prefix: &str) -> Result<MemberRole, sqlx::Error> {
    Ok(MemberRole {
        id: row.try_get(format!("{prefix}id").as_str())?,
        space_id: row.try_get(format!("{prefix}space_id").as_str())?,
        member_id: row.try_get(format!("{prefix}member_id").as_str())?,
        role_id: row.try_get(format!("{prefix}role_id").as_str())?,
        audit: super::audit::create(row, prefix)?,
    })
}
