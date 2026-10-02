use serde::{Deserialize, Serialize};
use time::OffsetDateTime;

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct AuditCreate {
    #[serde(rename = "created_at", with = "time::serde::rfc3339")]
    pub at: OffsetDateTime,
    #[serde(rename = "created_by", skip_serializing_if = "Option::is_none")]
    pub by: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct AuditUpdate {
    #[serde(
        rename = "updated_at",
        default,
        with = "time::serde::rfc3339::option",
        skip_serializing_if = "Option::is_none"
    )]
    pub at: Option<OffsetDateTime>,
    #[serde(rename = "updated_by", skip_serializing_if = "Option::is_none")]
    pub by: Option<i64>,
}

#[derive(Debug, Clone, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct AuditDelete {
    #[serde(
        rename = "deleted_at",
        default,
        with = "time::serde::rfc3339::option",
        skip_serializing_if = "Option::is_none"
    )]
    pub at: Option<OffsetDateTime>,
    #[serde(rename = "deleted_by", skip_serializing_if = "Option::is_none")]
    pub by: Option<i64>,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct AuditCreateUpdateDelete {
    #[serde(flatten)]
    pub create: AuditCreate,
    #[serde(flatten)]
    pub update: AuditUpdate,
    #[serde(flatten)]
    pub delete: AuditDelete,
}
