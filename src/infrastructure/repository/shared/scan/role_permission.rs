use mate_pgdt::sqlx::{self, Row, postgres::PgRow};

use crate::domain::models::RolePermission;

pub(crate) fn decode(row: &PgRow, prefix: &str) -> Result<RolePermission, sqlx::Error> {
    Ok(RolePermission {
        id: row.try_get(format!("{prefix}id").as_str())?,
        role_id: row.try_get(format!("{prefix}role_id").as_str())?,
        permission_id: row.try_get(format!("{prefix}permission_id").as_str())?,
        audit: super::audit::create(row, prefix)?,
    })
}
