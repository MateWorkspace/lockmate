use serde::{Deserialize, Serialize};

/// Access to one space membership; grants contain role and permission slugs.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct AccessClaims {
    pub user_id: i64,
    pub space_id: i64,
    pub member_id: i64,
    pub name: String,
    pub username: String,
    pub roles: Vec<String>,
    pub permissions: Vec<String>,
}

/// Identity and scope for refreshing within one space membership.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct RefreshClaims {
    pub user_id: i64,
    pub space_id: i64,
    pub member_id: i64,
    pub name: String,
    pub username: String,
}
