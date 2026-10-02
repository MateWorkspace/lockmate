#[derive(Debug, thiserror::Error, Clone, Copy, PartialEq, Eq)]
pub enum ValidatorError {
    #[error("invalid arguments")]
    BadArgs,

    /// Permission

    #[error("permission name is invalid")]
    PermissionNameInvalid,

    #[error("permission name is too long")]
    PermissionNameTooLong,

    #[error("permission name is too short")]
    PermissionNameTooShort,

    #[error("permission description is too long")]
    PermissionDescTooLong,

    /// Role

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

    #[error("api key name is invalid")]
    UserApiKeyNameInvalid,

    #[error("api key name is too long")]
    UserApiKeyNameTooLong,

    #[error("api key name is too short")]
    UserApiKeyNameTooShort,

    #[error("api key description is invalid")]
    UserApiKeyDescInvalid,

    #[error("api key description is too long")]
    UserApiKeyDescTooLong,
}

impl ValidatorError {
    pub const fn code(&self) -> &'static str {
        match self {
            Self::BadArgs => "BAD_ARGS",

            Self::PermissionNameInvalid => "PERMISSION_NAME_INVALID",
            Self::PermissionNameTooLong => "PERMISSION_NAME_TOO_LONG",
            Self::PermissionNameTooShort => "PERMISSION_NAME_TOO_SHORT",
            Self::PermissionDescTooLong => "PERMISSION_DESC_TOO_LONG",

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
            Self::UserApiKeyNameInvalid => "API_KEY_NAME_INVALID",
            Self::UserApiKeyNameTooLong => "API_KEY_NAME_TOO_LONG",
            Self::UserApiKeyNameTooShort => "API_KEY_NAME_TOO_SHORT",
            Self::UserApiKeyDescInvalid => "API_KEY_DESC_INVALID",
            Self::UserApiKeyDescTooLong => "API_KEY_DESC_TOO_LONG",
        }
    }
}
