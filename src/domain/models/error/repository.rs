#[derive(Debug, thiserror::Error)]
pub enum RepositoryError {

    /// Generic errors.
    
    #[error("invalid arguments")]
    BadArgs,

    #[error("record was not found")]
    NotFound,

    #[error("record conflicts with existing data")]
    Conflict,

    #[error("invalid state")]
    BadState,

    #[error("repository operation timed out")]
    Timeout {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },

    #[error("repository operation failed")]
    Failure {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync + 'static>,
    },

    #[error("operation is not implemented")]
    Unimplemented,

    /// Specific errors.

    #[error("permission was not found")]
    PermissionNotFound,

    #[error("permission name conflicts with existing data")]
    PermissionNameConflict,

    #[error("role was not found")]
    RoleNotFound,

    #[error("role name conflicts with existing data")]
    RoleNameConflict,

    #[error("role permission was not found")]
    RolePermissionNotFound,

    #[error("role permission conflicts with existing data")]
    RolePermissionConflict,

    #[error("user was not found")]
    UserNotFound,

    #[error("user username conflicts with existing data")]
    UserUsernameConflict,

    #[error("user email conflicts with existing data")]
    UserEmailConflict,

    #[error("user phone conflicts with existing data")]
    UserPhoneConflict,

    #[error("api key was not found")]
    UserApiKeyNotFound,
}

impl RepositoryError {
    pub const fn code(&self) -> &'static str {
        match self {
            Self::BadArgs => "BAD_ARGS",
            Self::NotFound => "NOT_FOUND",
            Self::Conflict => "CONFLICT",
            Self::BadState => "BAD_STATE",
            Self::Timeout { .. } => "TIMEOUT",
            Self::Failure { .. } => "FAILURE",
            Self::Unimplemented => "UNIMPLEMENTED",
            Self::PermissionNotFound => "PERMISSION_NOT_FOUND",
            Self::PermissionNameConflict => "PERMISSION_NAME_CONFLICT",
            Self::RoleNotFound => "ROLE_NOT_FOUND",
            Self::RoleNameConflict => "ROLE_NAME_CONFLICT",
            Self::RolePermissionNotFound => "ROLE_PERMISSION_NOT_FOUND",
            Self::RolePermissionConflict => "ROLE_PERMISSION_CONFLICT",
            Self::UserNotFound => "USER_NOT_FOUND",
            Self::UserUsernameConflict => "USER_USERNAME_CONFLICT",
            Self::UserEmailConflict => "USER_EMAIL_CONFLICT",
            Self::UserPhoneConflict => "USER_PHONE_CONFLICT",
            Self::UserApiKeyNotFound => "API_KEY_NOT_FOUND",
        }
    }
}
