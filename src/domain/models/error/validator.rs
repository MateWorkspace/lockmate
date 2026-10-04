#[derive(Debug, thiserror::Error, Clone, Copy, PartialEq, Eq)]
pub enum ValidatorError {
    #[error("invalid arguments")]
    BadArgs,

    /// environment

    #[error("{key} is required")]
    EnvironmentRequired { key: &'static str },

    #[error("{key} {reason}")]
    EnvironmentInvalid {
        key: &'static str,
        reason: &'static str,
    },

    /// Space

    #[error("space slug is invalid")]
    SpaceSlugInvalid,

    #[error("space slug is too long")]
    SpaceSlugTooLong,

    #[error("space slug is too short")]
    SpaceSlugTooShort,

    #[error("space name is invalid")]
    SpaceNameInvalid,

    #[error("space name is too long")]
    SpaceNameTooLong,

    #[error("space name is too short")]
    SpaceNameTooShort,

    #[error("space description is too long")]
    SpaceDescTooLong,

    /// Permission

    #[error("permission slug is invalid")]
    PermissionSlugInvalid,

    #[error("permission slug is too long")]
    PermissionSlugTooLong,

    #[error("permission slug is too short")]
    PermissionSlugTooShort,

    #[error("permission name is invalid")]
    PermissionNameInvalid,

    #[error("permission name is too long")]
    PermissionNameTooLong,

    #[error("permission name is too short")]
    PermissionNameTooShort,

    #[error("permission description is too long")]
    PermissionDescTooLong,

    /// Role

    #[error("role slug is invalid")]
    RoleSlugInvalid,

    #[error("role slug is too long")]
    RoleSlugTooLong,

    #[error("role slug is too short")]
    RoleSlugTooShort,

    #[error("role name is invalid")]
    RoleNameInvalid,

    #[error("role name is too long")]
    RoleNameTooLong,

    #[error("role name is too short")]
    RoleNameTooShort,

    #[error("role description is too long")]
    RoleDescTooLong,

    /// User

    #[error("user name is invalid")]
    UserNameInvalid,

    #[error("user name is too long")]
    UserNameTooLong,

    #[error("user name is too short")]
    UserNameTooShort,

    #[error("user bio is too long")]
    UserBioTooLong,

    #[error("user username is invalid")]
    UserUsernameInvalid,

    #[error("user username is too long")]
    UserUsernameTooLong,

    #[error("user username is too short")]
    UserUsernameTooShort,

    #[error("user email is invalid")]
    UserEmailInvalid,

    #[error("user email is too long")]
    UserEmailTooLong,

    #[error("user email is too short")]
    UserEmailTooShort,

    #[error("user phone is invalid")]
    UserPhoneInvalid,

    #[error("user phone is too long")]
    UserPhoneTooLong,

    #[error("user phone is too short")]
    UserPhoneTooShort,

    #[error("user password is invalid")]
    UserPasswordInvalid,

    #[error("user password is too long")]
    UserPasswordTooLong,

    #[error("user password is too short")]
    UserPasswordTooShort,

    /// API key

    #[error("api key name is invalid")]
    ApiKeyNameInvalid,

    #[error("api key name is too long")]
    ApiKeyNameTooLong,

    #[error("api key name is too short")]
    ApiKeyNameTooShort,

    #[error("api key description is invalid")]
    ApiKeyDescInvalid,

    #[error("api key description is too long")]
    ApiKeyDescTooLong,
}

impl ValidatorError {
    pub const fn code(&self) -> &'static str {
        match self {
            Self::BadArgs => "BAD_ARGS",
            Self::EnvironmentRequired { .. } => "ENVIRONMENT_REQUIRED",
            Self::EnvironmentInvalid { .. } => "ENVIRONMENT_INVALID",
            Self::SpaceSlugInvalid => "SPACE_SLUG_INVALID",
            Self::SpaceSlugTooLong => "SPACE_SLUG_TOO_LONG",
            Self::SpaceSlugTooShort => "SPACE_SLUG_TOO_SHORT",
            Self::SpaceNameInvalid => "SPACE_NAME_INVALID",
            Self::SpaceNameTooLong => "SPACE_NAME_TOO_LONG",
            Self::SpaceNameTooShort => "SPACE_NAME_TOO_SHORT",
            Self::SpaceDescTooLong => "SPACE_DESC_TOO_LONG",
            Self::PermissionSlugInvalid => "PERMISSION_SLUG_INVALID",
            Self::PermissionSlugTooLong => "PERMISSION_SLUG_TOO_LONG",
            Self::PermissionSlugTooShort => "PERMISSION_SLUG_TOO_SHORT",
            Self::PermissionNameInvalid => "PERMISSION_NAME_INVALID",
            Self::PermissionNameTooLong => "PERMISSION_NAME_TOO_LONG",
            Self::PermissionNameTooShort => "PERMISSION_NAME_TOO_SHORT",
            Self::PermissionDescTooLong => "PERMISSION_DESC_TOO_LONG",
            Self::RoleSlugInvalid => "ROLE_SLUG_INVALID",
            Self::RoleSlugTooLong => "ROLE_SLUG_TOO_LONG",
            Self::RoleSlugTooShort => "ROLE_SLUG_TOO_SHORT",
            Self::RoleNameInvalid => "ROLE_NAME_INVALID",
            Self::RoleNameTooLong => "ROLE_NAME_TOO_LONG",
            Self::RoleNameTooShort => "ROLE_NAME_TOO_SHORT",
            Self::RoleDescTooLong => "ROLE_DESC_TOO_LONG",
            Self::UserNameInvalid => "USER_NAME_INVALID",
            Self::UserNameTooLong => "USER_NAME_TOO_LONG",
            Self::UserNameTooShort => "USER_NAME_TOO_SHORT",
            Self::UserBioTooLong => "USER_BIO_TOO_LONG",
            Self::UserUsernameInvalid => "USER_USERNAME_INVALID",
            Self::UserUsernameTooLong => "USER_USERNAME_TOO_LONG",
            Self::UserUsernameTooShort => "USER_USERNAME_TOO_SHORT",
            Self::UserEmailInvalid => "USER_EMAIL_INVALID",
            Self::UserEmailTooLong => "USER_EMAIL_TOO_LONG",
            Self::UserEmailTooShort => "USER_EMAIL_TOO_SHORT",
            Self::UserPhoneInvalid => "USER_PHONE_INVALID",
            Self::UserPhoneTooLong => "USER_PHONE_TOO_LONG",
            Self::UserPhoneTooShort => "USER_PHONE_TOO_SHORT",
            Self::UserPasswordInvalid => "USER_PASSWORD_INVALID",
            Self::UserPasswordTooLong => "USER_PASSWORD_TOO_LONG",
            Self::UserPasswordTooShort => "USER_PASSWORD_TOO_SHORT",
            Self::ApiKeyNameInvalid => "API_KEY_NAME_INVALID",
            Self::ApiKeyNameTooLong => "API_KEY_NAME_TOO_LONG",
            Self::ApiKeyNameTooShort => "API_KEY_NAME_TOO_SHORT",
            Self::ApiKeyDescInvalid => "API_KEY_DESC_INVALID",
            Self::ApiKeyDescTooLong => "API_KEY_DESC_TOO_LONG",
        }
    }
}
