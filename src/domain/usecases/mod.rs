#![doc = include_str!("README.md")]

pub mod auth;
pub mod management;
pub mod profile;
pub mod shared;

pub use shared::UsecaseFuture;
