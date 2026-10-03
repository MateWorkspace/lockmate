use std::sync::Arc;

use regex::Regex;

use crate::domain::{
    contracts::utility::{Logger, Validator},
    models::{AppContext, LoggerMeta, ValidatorError},
};

use super::utils::{validate_length, validate_pattern};

const NAME_MIN_LENGTH: usize = 3;
const NAME_MAX_LENGTH: usize = 100;
const TEXT_MAX_LENGTH: usize = 1000;
const USERNAME_MIN_LENGTH: usize = 3;
const USERNAME_MAX_LENGTH: usize = 32;
const EMAIL_MIN_LENGTH: usize = 3;
const EMAIL_MAX_LENGTH: usize = 254;
const PHONE_MIN_LENGTH: usize = 8;
const PHONE_MAX_LENGTH: usize = 15;
const PASSWORD_MIN_LENGTH: usize = 8;
const PASSWORD_MAX_BYTES: usize = 72;

pub struct RegexValidator {
    logger: Arc<dyn Logger>,
    permission_name: Regex,
    name: Regex,
    username: Regex,
    email: Regex,
    phone: Regex,
}

impl RegexValidator {
    pub fn new(logger: Arc<dyn Logger>) -> Self {
        Self {
            logger,
            permission_name: Regex::new(r"\A[A-Za-z][A-Za-z0-9_.:-]*\z")
                .expect("valid permission name pattern"),
            name: Regex::new(r"\A[\p{L}\p{N}][\p{L}\p{M}\p{N} ._'-]*\z")
                .expect("valid name pattern"),
            username: Regex::new(r"\A[A-Za-z0-9][A-Za-z0-9._-]*\z")
                .expect("valid username pattern"),
            email: Regex::new(r"\A[A-Za-z0-9.!#$%&'*+/=?^_`{|}~-]+@[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)+\z")
                .expect("valid email pattern"),
            phone: Regex::new(r"\A\+?[0-9]+\z").expect("valid phone pattern"),
        }
    }

    fn finish(
        &self,
        context: &AppContext,
        tag: &str,
        result: Result<(), ValidatorError>,
    ) -> Result<(), ValidatorError> {
        result.inspect_err(|error| {
            let meta = LoggerMeta::from([
                ("error_code".into(), error.code().into()),
                ("error".into(), error.to_string().into()),
            ]);
            self.logger.error(context, tag, &error.to_string(), &meta);
        })
    }
}

impl Validator for RegexValidator {
    fn permission_name(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/permission_name";
        self.finish(
            context,
            TAG,
            validate_pattern(
                value,
                NAME_MIN_LENGTH..=NAME_MAX_LENGTH,
                &self.permission_name,
                (
                    ValidatorError::PermissionNameInvalid,
                    ValidatorError::PermissionNameTooShort,
                    ValidatorError::PermissionNameTooLong,
                ),
            ),
        )
    }

    fn role_name(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/role_name";
        self.finish(
            context,
            TAG,
            validate_pattern(
                value,
                NAME_MIN_LENGTH..=NAME_MAX_LENGTH,
                &self.name,
                (
                    ValidatorError::RoleNameInvalid,
                    ValidatorError::RoleNameTooShort,
                    ValidatorError::RoleNameTooLong,
                ),
            ),
        )
    }

    fn user_name(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/user_name";
        self.finish(
            context,
            TAG,
            validate_pattern(
                value,
                NAME_MIN_LENGTH..=NAME_MAX_LENGTH,
                &self.name,
                (
                    ValidatorError::UserNameInvalid,
                    ValidatorError::UserNameTooShort,
                    ValidatorError::UserNameTooLong,
                ),
            ),
        )
    }

    fn user_username(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/user_username";
        self.finish(
            context,
            TAG,
            validate_pattern(
                value,
                USERNAME_MIN_LENGTH..=USERNAME_MAX_LENGTH,
                &self.username,
                (
                    ValidatorError::UserUsernameInvalid,
                    ValidatorError::UserUsernameTooShort,
                    ValidatorError::UserUsernameTooLong,
                ),
            ),
        )
    }

    fn api_key_name(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/api_key_name";
        self.finish(
            context,
            TAG,
            validate_pattern(
                value,
                NAME_MIN_LENGTH..=NAME_MAX_LENGTH,
                &self.name,
                (
                    ValidatorError::UserApiKeyNameInvalid,
                    ValidatorError::UserApiKeyNameTooShort,
                    ValidatorError::UserApiKeyNameTooLong,
                ),
            ),
        )
    }

    fn permission_desc(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/permission_desc";
        self.finish(
            context,
            TAG,
            validate_length(
                value,
                0..=TEXT_MAX_LENGTH,
                ValidatorError::BadArgs,
                ValidatorError::PermissionDescTooLong,
            ),
        )
    }

    fn role_desc(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/role_desc";
        self.finish(
            context,
            TAG,
            validate_length(
                value,
                0..=TEXT_MAX_LENGTH,
                ValidatorError::BadArgs,
                ValidatorError::RoleDescTooLong,
            ),
        )
    }

    fn user_bio(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/user_bio";
        self.finish(
            context,
            TAG,
            validate_length(
                value,
                0..=TEXT_MAX_LENGTH,
                ValidatorError::BadArgs,
                ValidatorError::UserBioTooLong,
            ),
        )
    }

    fn user_email(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/user_email";
        let result = validate_pattern(
            value,
            EMAIL_MIN_LENGTH..=EMAIL_MAX_LENGTH,
            &self.email,
            (
                ValidatorError::UserEmailInvalid,
                ValidatorError::UserEmailTooShort,
                ValidatorError::UserEmailTooLong,
            ),
        )
        .and_then(|()| {
            // The pattern allows only an unquoted ASCII address. Its local part must
            // also follow the dot-atom rules checked by Go's mail.ParseAddress.
            let local = value.split_once('@').expect("email pattern includes @").0;
            if local.starts_with('.') || local.ends_with('.') || local.contains("..") {
                return Err(ValidatorError::UserEmailInvalid);
            }
            Ok(())
        });
        self.finish(context, TAG, result)
    }

    fn user_phone(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/user_phone";
        let result = validate_length(
            value.strip_prefix('+').unwrap_or(value),
            PHONE_MIN_LENGTH..=PHONE_MAX_LENGTH,
            ValidatorError::UserPhoneTooShort,
            ValidatorError::UserPhoneTooLong,
        )
        .and_then(|()| {
            if !self.phone.is_match(value) {
                return Err(ValidatorError::UserPhoneInvalid);
            }
            Ok(())
        });
        self.finish(context, TAG, result)
    }

    fn user_password(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/user_password";
        let result = validate_length(
            value,
            PASSWORD_MIN_LENGTH..=PASSWORD_MAX_BYTES,
            ValidatorError::UserPasswordTooShort,
            ValidatorError::UserPasswordTooLong,
        )
        .and_then(|()| {
            if value.len() > PASSWORD_MAX_BYTES {
                return Err(ValidatorError::UserPasswordTooLong);
            }
            if value.chars().any(char::is_control) {
                return Err(ValidatorError::UserPasswordInvalid);
            }
            Ok(())
        });
        self.finish(context, TAG, result)
    }

    fn api_key_desc(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError> {
        const TAG: &str = "utility/validator/regex/api_key_desc";
        let result = validate_length(
            value,
            0..=TEXT_MAX_LENGTH,
            ValidatorError::BadArgs,
            ValidatorError::UserApiKeyDescTooLong,
        )
        .and_then(|()| {
            if value
                .chars()
                .any(|c| c.is_control() && !matches!(c, '\n' | '\r' | '\t'))
            {
                return Err(ValidatorError::UserApiKeyDescInvalid);
            }
            Ok(())
        });
        self.finish(context, TAG, result)
    }
}
