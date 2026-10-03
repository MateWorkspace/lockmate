use super::{SeedPermission, SeedRole, SeedUser};

#[derive(Clone, PartialEq, Eq)]
pub struct SeedData {
    pub permissions: Vec<SeedPermission>,
    pub roles: Vec<SeedRole>,
    pub users: Vec<SeedUser>,
}
