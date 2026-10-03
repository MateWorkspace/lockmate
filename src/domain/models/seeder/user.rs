use serde::Deserialize;

// Seed passwords are plaintext inputs to the password utility, never stored hashes.
#[derive(Clone, PartialEq, Eq, Deserialize)]
pub struct SeedUser {
    pub role_name: String,
    pub name: String,
    pub bio: Option<String>,
    pub username: String,
    pub password: String,
    pub avatar_path: Option<String>,
}
