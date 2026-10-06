use std::{future::Future, pin::Pin};

use crate::domain::models::UsecaseError;

pub type UsecaseFuture<'a, T> = Pin<Box<dyn Future<Output = Result<T, UsecaseError>> + Send + 'a>>;
