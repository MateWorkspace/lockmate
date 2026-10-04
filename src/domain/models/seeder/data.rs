use super::{SeedSpace, SeedSpaceMember, SeedUser};

#[derive(Clone, PartialEq, Eq)]
pub struct SeedData {
    pub spaces: Vec<SeedSpace>,
    pub users: Vec<SeedUser>,
    pub space_members: Vec<SeedSpaceMember>,
}
