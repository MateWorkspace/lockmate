//! Auth, management, and profile interfaces with shared safe response types.

pub mod auth;
pub mod management;
pub mod profile;
pub mod shared;

pub use shared::UsecaseFuture;
