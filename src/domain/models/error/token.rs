#[derive(Debug, thiserror::Error, Clone, Copy, PartialEq, Eq)]
pub enum TokenError {
    #[error("token is invalid")]
    Invalid,

    #[error("token has expired")]
    Expired,
}

impl TokenError {
    pub const fn code(&self) -> &'static str {
        match self {
            Self::Invalid => "TOKEN_INVALID",
            Self::Expired => "TOKEN_EXPIRED",
        }
    }
}
