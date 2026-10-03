use serde::Deserialize;

#[derive(Debug, Clone, PartialEq, Eq, Deserialize)]
pub struct SeedRole {
    pub name: String,
    pub description: String,
    pub is_default: bool,
    pub permissions: Vec<String>,
}
