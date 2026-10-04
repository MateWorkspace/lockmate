use std::sync::{Arc, Mutex};

use lockmate::{
    domain::{
        contracts::utility::{Logger, Validator},
        models::{AppContext, LoggerLevel, LoggerMeta, ValidatorError},
    },
    infrastructure::utility::{logger::JsonLogger, validator::RegexValidator},
};

type Check = fn(&dyn Validator, &AppContext, &str) -> Result<(), ValidatorError>;

fn new_validator() -> RegexValidator {
    RegexValidator::new(Arc::new(JsonLogger::with_writer(
        std::io::sink(),
        LoggerLevel::None,
    )))
}

#[test]
fn name_and_username_rules_preserve_field_errors_and_length_precedence() {
    let validator: Arc<dyn Validator> = Arc::new(new_validator());
    let context = AppContext::default();
    let cases: [(Check, usize, ValidatorError, ValidatorError, ValidatorError); 6] = [
        (
            |v, c, s| v.space_name(c, s),
            100,
            ValidatorError::SpaceNameTooShort,
            ValidatorError::SpaceNameTooLong,
            ValidatorError::SpaceNameInvalid,
        ),
        (
            |v, c, s| v.permission_name(c, s),
            100,
            ValidatorError::PermissionNameTooShort,
            ValidatorError::PermissionNameTooLong,
            ValidatorError::PermissionNameInvalid,
        ),
        (
            |v, c, s| v.role_name(c, s),
            100,
            ValidatorError::RoleNameTooShort,
            ValidatorError::RoleNameTooLong,
            ValidatorError::RoleNameInvalid,
        ),
        (
            |v, c, s| v.user_name(c, s),
            100,
            ValidatorError::UserNameTooShort,
            ValidatorError::UserNameTooLong,
            ValidatorError::UserNameInvalid,
        ),
        (
            |v, c, s| v.api_key_name(c, s),
            100,
            ValidatorError::ApiKeyNameTooShort,
            ValidatorError::ApiKeyNameTooLong,
            ValidatorError::ApiKeyNameInvalid,
        ),
        (
            |v, c, s| v.user_username(c, s),
            32,
            ValidatorError::UserUsernameTooShort,
            ValidatorError::UserUsernameTooLong,
            ValidatorError::UserUsernameInvalid,
        ),
    ];
    for (check, max, too_short, too_long, invalid) in cases {
        for value in ["", "ab", "!"] {
            assert_eq!(check(validator.as_ref(), &context, value), Err(too_short));
        }
        assert_eq!(check(validator.as_ref(), &context, "abc"), Ok(()));
        assert_eq!(
            check(validator.as_ref(), &context, &"a".repeat(max)),
            Ok(())
        );
        assert_eq!(
            check(validator.as_ref(), &context, &"!".repeat(max + 1)),
            Err(too_long)
        );
        for value in [" abc", "abc ", "abc\n", "a/b", "_abc", "a😀b"] {
            assert_eq!(check(validator.as_ref(), &context, value), Err(invalid));
        }
    }
    validator
        .permission_slug(&context, "user:read.all_items-1")
        .unwrap();
    assert_eq!(
        validator.permission_slug(&context, "1user"),
        Err(ValidatorError::PermissionSlugInvalid)
    );
    assert_eq!(
        validator.permission_slug(&context, "用户名"),
        Err(ValidatorError::PermissionSlugInvalid)
    );
    validator.user_username(&context, "1user.name-_9").unwrap();
    assert_eq!(
        validator.user_username(&context, "用户名"),
        Err(ValidatorError::UserUsernameInvalid)
    );
}

#[test]
fn display_names_support_unicode_letters_numbers_and_combining_marks() {
    let validator = new_validator();
    let context = AppContext::default();
    let checks: [Check; 5] = [
        |v, c, s| v.space_name(c, s),
        |v, c, s| v.permission_name(c, s),
        |v, c, s| v.role_name(c, s),
        |v, c, s| v.user_name(c, s),
        |v, c, s| v.api_key_name(c, s),
    ];
    for check in checks {
        for value in [
            "用户名",
            "Élodie O'Connor",
            "A\u{301}b",
            "123",
            "ⅧTest",
            "Main_key.v2-1",
            &"界".repeat(100),
        ] {
            check(&validator, &context, value).unwrap();
        }
    }
    assert_eq!(
        validator.user_name(&context, "界界"),
        Err(ValidatorError::UserNameTooShort)
    );
    assert_eq!(
        validator.user_name(&context, &"界".repeat(101)),
        Err(ValidatorError::UserNameTooLong)
    );
    assert_eq!(
        validator.role_name(&context, "\u{301}ab"),
        Err(ValidatorError::RoleNameInvalid)
    );
}

#[test]
fn optional_text_counts_characters_and_api_key_description_filters_controls() {
    let validator = new_validator();
    let context = AppContext::default();
    let checks: [(Check, ValidatorError); 5] = [
        (
            |v, c, s| v.space_desc(c, s),
            ValidatorError::SpaceDescTooLong,
        ),
        (
            |v, c, s| v.permission_desc(c, s),
            ValidatorError::PermissionDescTooLong,
        ),
        (|v, c, s| v.role_desc(c, s), ValidatorError::RoleDescTooLong),
        (|v, c, s| v.user_bio(c, s), ValidatorError::UserBioTooLong),
        (
            |v, c, s| v.api_key_desc(c, s),
            ValidatorError::ApiKeyDescTooLong,
        ),
    ];
    for (check, too_long) in checks {
        for value in ["", "\n\r\t", &"界".repeat(1000)] {
            check(&validator, &context, value).unwrap();
        }
        assert_eq!(
            check(&validator, &context, &"界".repeat(1001)),
            Err(too_long)
        );
    }
    for value in [
        "private\0description",
        "private\u{7f}description",
        "private\u{85}description",
    ] {
        assert_eq!(
            validator.api_key_desc(&context, value),
            Err(ValidatorError::ApiKeyDescInvalid)
        );
        validator.permission_desc(&context, value).unwrap();
        validator.role_desc(&context, value).unwrap();
        validator.user_bio(&context, value).unwrap();
    }
}

#[test]
fn email_requires_plain_ascii_address_and_valid_dot_atoms_and_domain_labels() {
    let validator = new_validator();
    let context = AppContext::default();
    for value in [
        "a@b.co",
        "first.last+tag@example.com",
        "!#$%&'*+/=?^_`{|}~-@sub-domain.example.com",
    ] {
        validator.user_email(&context, value).unwrap();
    }
    let max = format!(
        "{}@{}.{}.{}",
        "a".repeat(64),
        "b".repeat(63),
        "c".repeat(63),
        "d".repeat(61)
    );
    assert_eq!(max.len(), 254);
    validator.user_email(&context, &max).unwrap();
    assert_eq!(
        validator.user_email(&context, &(max + "d")),
        Err(ValidatorError::UserEmailTooLong)
    );
    assert_eq!(
        validator.user_email(&context, "a@"),
        Err(ValidatorError::UserEmailTooShort)
    );
    for value in [
        ".user@example.com",
        "user.@example.com",
        "us..er@example.com",
        "User <user@example.com>",
        "\"user\"@example.com",
        "user@example",
        "user@-example.com",
        "user@example-.com",
        "user@ex_ample.com",
        "user@example..com",
        "user@example.com.",
        "用户@example.com",
        "user@例子.com",
        " user@example.com",
        "user@example.com\n",
    ] {
        assert_eq!(
            validator.user_email(&context, value),
            Err(ValidatorError::UserEmailInvalid),
            "{value:?}"
        );
    }
    assert_eq!(
        validator.user_email(&context, &format!("user@{}.com", "a".repeat(64))),
        Err(ValidatorError::UserEmailInvalid)
    );
}

#[test]
fn phone_counts_digits_without_plus_and_rejects_non_ascii_or_formatting() {
    let validator = new_validator();
    let context = AppContext::default();
    for value in [
        "12345678",
        "+12345678",
        "123456789012345",
        "+123456789012345",
    ] {
        validator.user_phone(&context, value).unwrap();
    }
    for value in ["", "+", "1234567", "+1234567"] {
        assert_eq!(
            validator.user_phone(&context, value),
            Err(ValidatorError::UserPhoneTooShort)
        );
    }
    for value in ["1234567890123456", "+1234567890123456"] {
        assert_eq!(
            validator.user_phone(&context, value),
            Err(ValidatorError::UserPhoneTooLong)
        );
    }
    for value in [
        "1234-5678",
        "1234 5678",
        "++12345678",
        "１２３４５６７８",
        "1234567\n",
    ] {
        assert_eq!(
            validator.user_phone(&context, value),
            Err(ValidatorError::UserPhoneInvalid)
        );
    }
}

#[test]
fn password_counts_characters_for_minimum_and_utf8_bytes_for_bcrypt_limit() {
    let validator = new_validator();
    let context = AppContext::default();
    for value in [
        "12345678",
        "        ",
        "😀😀😀😀😀😀😀😀",
        &"a".repeat(72),
        &"界".repeat(24),
    ] {
        validator.user_password(&context, value).unwrap();
    }
    for value in ["", "1234567", "界界界界界界界"] {
        assert_eq!(
            validator.user_password(&context, value),
            Err(ValidatorError::UserPasswordTooShort)
        );
    }
    for value in [&"a".repeat(73), &"界".repeat(25), &"😀".repeat(19)] {
        assert_eq!(
            validator.user_password(&context, value),
            Err(ValidatorError::UserPasswordTooLong)
        );
    }
    for value in [
        "private\0password",
        "private\npassword",
        "private\tpassword",
        "private\u{7f}password",
        "private\u{85}password",
    ] {
        assert_eq!(
            validator.user_password(&context, value),
            Err(ValidatorError::UserPasswordInvalid)
        );
    }
}

struct LogEntry {
    context: AppContext,
    level: LoggerLevel,
    tag: String,
    message: String,
    meta: LoggerMeta,
}

#[derive(Default)]
struct RecordingLogger {
    entries: Mutex<Vec<LogEntry>>,
}

impl Logger for RecordingLogger {
    fn log(
        &self,
        context: &AppContext,
        level: LoggerLevel,
        tag: &str,
        message: &str,
        meta: &LoggerMeta,
    ) {
        self.entries.lock().unwrap().push(LogEntry {
            context: context.clone(),
            level,
            tag: tag.to_owned(),
            message: message.to_owned(),
            meta: meta.clone(),
        });
    }
}

#[test]
fn each_validation_failure_logs_once_with_context_and_without_submitted_values() {
    let logger = Arc::new(RecordingLogger::default());
    let validator = RegexValidator::new(logger.clone());
    let context = AppContext {
        actor: Some("admin".into()),
        trace_id: Some(uuid::Uuid::from_u128(123)),
        ..AppContext::default()
    };
    validator
        .user_password(&context, "private password")
        .unwrap();
    assert!(logger.entries.lock().unwrap().is_empty());
    let long_text = "private text".repeat(100);
    let errors = [
        (
            "space_slug",
            validator.space_slug(&context, "private/space"),
        ),
        ("space_name", validator.space_name(&context, "private/name")),
        ("space_desc", validator.space_desc(&context, &long_text)),
        ("role_slug", validator.role_slug(&context, "private/role")),
        (
            "permission_slug",
            validator.permission_slug(&context, "private/permission"),
        ),
        (
            "permission_name",
            validator.permission_name(&context, "private/name"),
        ),
        (
            "permission_desc",
            validator.permission_desc(&context, &long_text),
        ),
        ("role_name", validator.role_name(&context, "private/name")),
        ("role_desc", validator.role_desc(&context, &long_text)),
        ("user_name", validator.user_name(&context, "private/name")),
        ("user_bio", validator.user_bio(&context, &long_text)),
        (
            "user_username",
            validator.user_username(&context, "private/name"),
        ),
        (
            "user_email",
            validator.user_email(&context, "private-email"),
        ),
        (
            "user_phone",
            validator.user_phone(&context, "private-phone"),
        ),
        (
            "user_password",
            validator.user_password(&context, "private\0password"),
        ),
        (
            "api_key_name",
            validator.api_key_name(&context, "private/name"),
        ),
        (
            "api_key_desc",
            validator.api_key_desc(&context, "private\0description"),
        ),
    ];
    let entries = logger.entries.lock().unwrap();
    assert_eq!(entries.len(), errors.len());
    for (entry, (method, result)) in entries.iter().zip(errors) {
        let error = result.unwrap_err();
        assert_eq!(entry.context, context);
        assert_eq!(entry.level, LoggerLevel::Error);
        assert_eq!(entry.tag, format!("utility/validator/regex/{method}"));
        assert_eq!(entry.message, error.to_string());
        assert_eq!(entry.meta["error_code"], error.code().into());
        assert_eq!(entry.meta["error"], error.to_string().into());
        assert!(!format!("{} {:?}", entry.message, entry.meta).contains("private"));
    }
}

#[test]
fn slug_fields_enforce_length_case_and_entity_specific_syntax() {
    let validator = new_validator();
    let context = AppContext::default();
    let cases: [(Check, ValidatorError, ValidatorError, ValidatorError, &str); 3] = [
        (
            |v, c, s| v.space_slug(c, s),
            ValidatorError::SpaceSlugTooShort,
            ValidatorError::SpaceSlugTooLong,
            ValidatorError::SpaceSlugInvalid,
            "SPACE_SLUG_INVALID",
        ),
        (
            |v, c, s| v.role_slug(c, s),
            ValidatorError::RoleSlugTooShort,
            ValidatorError::RoleSlugTooLong,
            ValidatorError::RoleSlugInvalid,
            "ROLE_SLUG_INVALID",
        ),
        (
            |v, c, s| v.permission_slug(c, s),
            ValidatorError::PermissionSlugTooShort,
            ValidatorError::PermissionSlugTooLong,
            ValidatorError::PermissionSlugInvalid,
            "PERMISSION_SLUG_INVALID",
        ),
    ];
    for (check, short, long, invalid, code) in cases {
        assert_eq!(invalid.code(), code);
        for value in ["", "ab", "!"] {
            assert_eq!(check(&validator, &context, value), Err(short));
        }
        check(&validator, &context, "abc").unwrap();
        check(&validator, &context, &"a".repeat(100)).unwrap();
        assert_eq!(check(&validator, &context, &"!".repeat(101)), Err(long));
        for value in [
            "ABC",
            "someThing",
            "1abc",
            " abc",
            "abc ",
            "abc\n",
            "a/b",
            "用户名",
            "a😀b",
        ] {
            assert_eq!(check(&validator, &context, value), Err(invalid));
        }
    }
    for check in [
        |v: &dyn Validator, c: &AppContext, s: &str| v.space_slug(c, s),
        |v: &dyn Validator, c: &AppContext, s: &str| v.role_slug(c, s),
    ] {
        for value in ["taskmate", "project-v2", "some-space-42"] {
            check(&validator, &context, value).unwrap();
        }
        for value in ["-abc", "abc-", "abc--def", "abc_def", "abc.def", "abc:def"] {
            assert!(check(&validator, &context, value).is_err());
        }
    }
    for value in [
        "profile:get",
        "profile_security:set",
        "settings.read",
        "user:read.all_items-1",
    ] {
        validator.permission_slug(&context, value).unwrap();
    }
    assert_eq!(
        ValidatorError::ApiKeyNameInvalid.code(),
        "API_KEY_NAME_INVALID"
    );
    assert_eq!(
        ValidatorError::ApiKeyDescTooLong.code(),
        "API_KEY_DESC_TOO_LONG"
    );
}
