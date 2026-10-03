use std::{future::Future, pin::Pin};

use crate::domain::models::{AppContext, TransactorError};

pub type TransactionFuture<'a> =
    Pin<Box<dyn Future<Output = Result<(), TransactorError>> + Send + 'a>>;

pub type TransactionCallback<'a> = Box<dyn FnOnce(AppContext) -> TransactionFuture<'a> + Send + 'a>;

pub trait Transactor: Send + Sync {
    /// Runs the callback in a transaction, committing success and rolling back errors.
    /// Pass the derived context to every repository call. Nested calls reuse the
    /// transaction; propagate their errors when the outer owner should roll back.
    /// Convert repository errors with `?`; callback causes are preserved separately
    /// from transaction-driver failures.
    ///
    /// ```no_run
    /// use lockmate::domain::{
    ///     contracts::utility::Transactor,
    ///     models::{AppContext, RepositoryError, TransactorError},
    /// };
    ///
    /// type RepositoryFuture<'a> = std::pin::Pin<Box<
    ///     dyn std::future::Future<Output = Result<(), RepositoryError>> + Send + 'a
    /// >>;
    ///
    /// trait Repository: Send + Sync {
    ///     fn create<'a>(&'a self, context: &'a AppContext) -> RepositoryFuture<'a>;
    /// }
    ///
    /// async fn create(
    ///     transactor: &dyn Transactor,
    ///     repository: &dyn Repository,
    ///     context: &AppContext,
    /// ) -> Result<(), TransactorError> {
    ///     transactor.with_tx(context, Box::new(move |context| Box::pin(async move {
    ///         repository.create(&context).await?;
    ///         Ok(())
    ///     }))).await
    /// }
    /// ```
    fn with_tx<'a>(
        &'a self,
        context: &'a AppContext,
        operation: TransactionCallback<'a>,
    ) -> TransactionFuture<'a>;
}
