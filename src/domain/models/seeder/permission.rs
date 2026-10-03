use serde::Deserialize;

#[derive(Debug, Clone, PartialEq, Eq, Deserialize)]
pub struct SeedPermission {
    pub name: String,
    pub description: String,
}
