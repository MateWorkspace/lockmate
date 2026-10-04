use std::{future::Future, pin::Pin};

use crate::domain::models::CachingError;

pub type CachingFuture<'a, T> = Pin<Box<dyn Future<Output = Result<T, CachingError>> + Send + 'a>>;
