use serde::Deserialize;

use super::{SeedPermission, SeedRole};

#[derive(Debug, Clone, PartialEq, Eq, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SeedSpace {
    pub slug: String,
    pub name: String,
    pub description: String,
    #[serde(default = "super::default_active")]
    pub is_active: bool,
    pub permissions: Vec<SeedPermission>,
    pub roles: Vec<SeedRole>,
}
