use super::RepositoryError;

#[derive(Debug, thiserror::Error)]
pub enum TransactorError {
    #[error("transaction operation timed out")]
    Timeout {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },

    #[error("transaction operation failed")]
    Failure {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },

    #[error("transaction callback failed")]
    Callback {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },
}

impl TransactorError {
    pub const fn code(&self) -> &'static str {
        match self {
            Self::Timeout { .. } => "TIMEOUT",
            Self::Failure { .. } | Self::Callback { .. } => "FAILURE",
        }
    }
}

impl From<RepositoryError> for TransactorError {
    fn from(source: RepositoryError) -> Self {
        Self::Callback {
            source: Box::new(source),
        }
    }
}
