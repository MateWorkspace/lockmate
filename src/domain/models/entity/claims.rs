use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct AccessClaims {
    pub user_id: i64,
    pub name: String,
    pub username: String,
    pub role: String,
    pub permissions: Vec<String>,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct RefreshClaims {
    pub user_id: i64,
    pub name: String,
    pub username: String,
}
