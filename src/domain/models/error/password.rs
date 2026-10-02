#[derive(Debug, thiserror::Error)]
pub enum PasswordError {
    #[error("password does not match")]
    Mismatch,

    #[error("failed to hash password")]
    HashFailure {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },

    #[error("failed to compare password")]
    CompareFailure {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },
}

impl PasswordError {
    pub const fn code(&self) -> &'static str {
        match self {
            Self::Mismatch => "UNAUTHORIZED",
            Self::HashFailure { .. } | Self::CompareFailure { .. } => "FAILURE",
        }
    }
}
