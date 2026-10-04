use serde::Deserialize;

#[derive(Debug, Clone, PartialEq, Eq, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SeedRole {
    pub slug: String,
    pub name: String,
    pub description: String,
    pub is_default: bool,
    /// Permission slugs within this role's space.
    pub permissions: Vec<String>,
}
