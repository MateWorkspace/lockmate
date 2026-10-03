use std::sync::Arc;

use mate_pgdt::sqlx;

use crate::domain::{
    contracts::utility::{Logger, TransactionCallback, TransactionFuture, Transactor},
    models::{AppContext, LoggerMeta, TransactorError},
};

pub struct MatePgdtTransactor {
    transactor: mate_pgdt::Transactor,
    logger: Arc<dyn Logger>,
}

impl MatePgdtTransactor {
    pub fn new(transactor: mate_pgdt::Transactor, logger: Arc<dyn Logger>) -> Self {
        Self { transactor, logger }
    }

    fn driver_error(&self, context: &AppContext, source: sqlx::Error) -> TransactorError {
        const TAG: &str = "utility/transactor/mate_pgdt/with_tx";

        let mut meta = LoggerMeta::new();

        if let Some(error) = source.as_database_error() {
            if let Some(code) = error.code() {
                meta.insert("sql_state".into(), code.into_owned().into());
            }
            if let Some(constraint) = error.constraint() {
                meta.insert("constraint".into(), constraint.into());
            }
        }

        let timeout = matches!(source, sqlx::Error::PoolTimedOut)
            || matches!(&source, sqlx::Error::Io(error) if error.kind() == std::io::ErrorKind::TimedOut);

        let error = if timeout {
            TransactorError::Timeout {
                source: Box::new(source),
            }
        } else {
            TransactorError::Failure {
                source: Box::new(source),
            }
        };

        meta.insert("error_code".into(), error.code().into());
        self.logger.error(context, TAG, &error.to_string(), &meta);
        error
    }
}

impl Transactor for MatePgdtTransactor {
    fn with_tx<'a>(
        &'a self,
        context: &'a AppContext,
        operation: TransactionCallback<'a>,
    ) -> TransactionFuture<'a> {
        Box::pin(async move {
            self.transactor
                .with_tx(context, move |context| async move {
                    operation(context).await.map_err(TransactionError::Callback)
                })
                .await
                .map_err(|error| match error {
                    TransactionError::Callback(error) => error,
                    TransactionError::Driver(source) => self.driver_error(context, source),
                })
        })
    }
}

enum TransactionError {
    Callback(TransactorError),
    Driver(sqlx::Error),
}

impl From<sqlx::Error> for TransactionError {
    fn from(error: sqlx::Error) -> Self {
        Self::Driver(error)
    }
}
