pub mod data;
pub mod permission;
pub mod role;
pub mod space;
pub mod space_member;
pub mod user;

pub use data::SeedData;
pub use permission::SeedPermission;
pub use role::SeedRole;
pub use space::SeedSpace;
pub use space_member::SeedSpaceMember;
pub use user::SeedUser;

const fn default_active() -> bool {
    true
}
