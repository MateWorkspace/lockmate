use serde::Deserialize;

#[derive(Debug, Clone, PartialEq, Eq, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SeedPermission {
    pub slug: String,
    pub name: String,
    pub description: String,
}
