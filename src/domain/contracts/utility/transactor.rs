use std::{future::Future, pin::Pin};

use crate::domain::models::{AppContext, TransactorError};

pub trait Transactor: Send + Sync {
    fn with_tx<'a>(
        &'a self,
        context: &'a AppContext,
        operation: TransactionCallback<'a>,
    ) -> TransactionFuture<'a>;
}

pub type TransactionFuture<'a> =
    Pin<Box<dyn Future<Output = Result<(), TransactorError>> + Send + 'a>>;

pub type TransactionCallback<'a> = Box<dyn FnOnce(AppContext) -> TransactionFuture<'a> + Send + 'a>;
