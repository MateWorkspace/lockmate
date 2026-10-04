use serde::Deserialize;

#[derive(Debug, Clone, PartialEq, Eq, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SeedSpaceMember {
    pub space_slug: String,
    pub username: String,
    #[serde(default = "super::default_active")]
    pub is_active: bool,
    /// Additional role slugs; membership creation already assigns the default role.
    #[serde(default)]
    pub roles: Vec<String>,
}
