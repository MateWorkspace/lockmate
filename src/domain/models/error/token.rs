#[derive(Debug, thiserror::Error)]
pub enum TokenError {
    #[error("token is required")]
    BadArgs,

    #[error("token is invalid")]
    Invalid,

    #[error("token has expired")]
    Expired,

    #[error("token operation failed")]
    Failure {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },
}

impl TokenError {
    pub const fn code(&self) -> &'static str {
        match self {
            Self::BadArgs => "BAD_ARGS",
            Self::Invalid => "TOKEN_INVALID",
            Self::Expired => "TOKEN_EXPIRED",
            Self::Failure { .. } => "FAILURE",
        }
    }
}
