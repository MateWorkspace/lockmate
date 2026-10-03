use mate_pgdt::sqlx::{self, Row, postgres::PgRow};

use crate::domain::models::{AuditCreate, AuditCreateUpdateDelete, AuditDelete, AuditUpdate};

pub(super) fn create(row: &PgRow, prefix: &str) -> Result<AuditCreate, sqlx::Error> {
    Ok(AuditCreate {
        at: row.try_get(format!("{prefix}created_at").as_str())?,
        by: row.try_get(format!("{prefix}created_by").as_str())?,
    })
}

pub(super) fn full(row: &PgRow, prefix: &str) -> Result<AuditCreateUpdateDelete, sqlx::Error> {
    Ok(AuditCreateUpdateDelete {
        create: create(row, prefix)?,
        update: AuditUpdate {
            at: row.try_get(format!("{prefix}updated_at").as_str())?,
            by: row.try_get(format!("{prefix}updated_by").as_str())?,
        },
        delete: AuditDelete {
            at: row.try_get(format!("{prefix}deleted_at").as_str())?,
            by: row.try_get(format!("{prefix}deleted_by").as_str())?,
        },
    })
}
