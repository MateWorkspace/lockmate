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

    #[error("space was not found")]
    SpaceNotFound,

    #[error("space slug conflicts with existing data")]
    SpaceSlugConflict,

    #[error("space member was not found")]
    SpaceMemberNotFound,

    #[error("space member conflicts with existing data")]
    SpaceMemberConflict,

    #[error("member role was not found")]
    MemberRoleNotFound,

    #[error("member role conflicts with existing data")]
    MemberRoleConflict,

    #[error("permission was not found")]
    PermissionNotFound,

    #[error("permission slug conflicts with existing data")]
    PermissionSlugConflict,

    #[error("role was not found")]
    RoleNotFound,

    #[error("role slug conflicts with existing data")]
    RoleSlugConflict,

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
    ApiKeyNotFound,
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
            Self::SpaceNotFound => "SPACE_NOT_FOUND",
            Self::SpaceSlugConflict => "SPACE_SLUG_CONFLICT",
            Self::SpaceMemberNotFound => "SPACE_MEMBER_NOT_FOUND",
            Self::SpaceMemberConflict => "SPACE_MEMBER_CONFLICT",
            Self::MemberRoleNotFound => "MEMBER_ROLE_NOT_FOUND",
            Self::MemberRoleConflict => "MEMBER_ROLE_CONFLICT",
            Self::PermissionNotFound => "PERMISSION_NOT_FOUND",
            Self::PermissionSlugConflict => "PERMISSION_SLUG_CONFLICT",
            Self::RoleNotFound => "ROLE_NOT_FOUND",
            Self::RoleSlugConflict => "ROLE_SLUG_CONFLICT",
            Self::RolePermissionNotFound => "ROLE_PERMISSION_NOT_FOUND",
            Self::RolePermissionConflict => "ROLE_PERMISSION_CONFLICT",
            Self::UserNotFound => "USER_NOT_FOUND",
            Self::UserUsernameConflict => "USER_USERNAME_CONFLICT",
            Self::UserEmailConflict => "USER_EMAIL_CONFLICT",
            Self::UserPhoneConflict => "USER_PHONE_CONFLICT",
            Self::ApiKeyNotFound => "API_KEY_NOT_FOUND",
        }
    }
}
