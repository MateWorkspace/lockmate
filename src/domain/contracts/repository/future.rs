use std::{future::Future, pin::Pin};

use crate::domain::models::RepositoryError;

pub type RepositoryFuture<'a, T> =
    Pin<Box<dyn Future<Output = Result<T, RepositoryError>> + Send + 'a>>;
