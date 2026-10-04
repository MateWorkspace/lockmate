#[derive(Debug, thiserror::Error)]
pub enum CachingError {
    #[error("invalid arguments")]
    BadArgs,

    #[error("invalid state")]
    BadState,

    #[error("caching operation timed out")]
    Timeout {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },

    #[error("caching operation failed")]
    Failure {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },
}

impl CachingError {
    pub const fn code(&self) -> &'static str {
        match self {
            Self::BadArgs => "BAD_ARGS",
            Self::BadState => "BAD_STATE",
            Self::Timeout { .. } => "TIMEOUT",
            Self::Failure { .. } => "FAILURE",
        }
    }
}
