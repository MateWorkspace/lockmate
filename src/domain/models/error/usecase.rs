use super::{
    ApiKeyError, PasswordError, RepositoryError, TokenError, TransactorError, ValidatorError,
};

#[derive(Debug, thiserror::Error)]
pub enum UsecaseError {
    #[error("invalid arguments")]
    BadArgs,

    #[error("invalid state")]
    BadState,

    #[error("authentication failed")]
    Unauthorized {
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync + 'static>>,
    },

    #[error("access is forbidden")]
    Forbidden,

    #[error("{0}")]
    Validator(#[from] ValidatorError),

    #[error("{0}")]
    Repository(#[from] RepositoryError),

    #[error("{0}")]
    Password(#[from] PasswordError),

    #[error("{0}")]
    Token(#[from] TokenError),

    #[error("{0}")]
    ApiKey(#[from] ApiKeyError),

    #[error("{0}")]
    Transactor(#[from] TransactorError),
}

impl UsecaseError {
    pub const fn code(&self) -> &'static str {
        match self {
            Self::BadArgs => "BAD_ARGS",
            Self::BadState => "BAD_STATE",
            Self::Unauthorized { .. } => "UNAUTHORIZED",
            Self::Forbidden => "FORBIDDEN",
            Self::Validator(source) => source.code(),
            Self::Repository(source) => source.code(),
            Self::Password(source) => source.code(),
            Self::Token(source) => source.code(),
            Self::ApiKey(source) => source.code(),
            Self::Transactor(source) => source.code(),
        }
    }
}
