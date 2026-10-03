#[derive(Debug, thiserror::Error)]
pub enum ApiKeyError {
    #[error("api key is required")]
    BadArgs,

    #[error("api key is invalid")]
    Invalid,

    #[error("failed to generate api key")]
    Failure {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },
}

impl ApiKeyError {
    pub const fn code(&self) -> &'static str {
        match self {
            Self::BadArgs => "BAD_ARGS",
            Self::Invalid => "API_KEY_INVALID",
            Self::Failure { .. } => "FAILURE",
        }
    }
}
