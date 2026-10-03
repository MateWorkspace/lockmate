use std::{
    error::Error as _,
    sync::{Arc, Mutex},
};

use lockmate::{
    domain::{
        contracts::utility::{Logger, Password},
        models::{AppContext, LoggerLevel, LoggerMeta, PasswordError},
    },
    infrastructure::utility::{logger::JsonLogger, password::BcryptPassword},
};

fn new_password(cost: u32) -> BcryptPassword {
    let logger = Arc::new(JsonLogger::with_writer(std::io::sink(), LoggerLevel::Error));
    BcryptPassword::new(logger, cost)
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
            tag: tag.into(),
            message: message.into(),
            meta: meta.clone(),
        });
    }
}

#[test]
fn errors_are_logged_once_with_request_context_and_without_credentials() {
    let logger = Arc::new(RecordingLogger::default());
    let password = BcryptPassword::new(logger.clone(), 4);
    let context = AppContext {
        actor: Some("admin".into()),
        trace_id: Some(uuid::Uuid::from_u128(42)),
    };
    let secret = "private password";
    let hash = password.hash(&context, secret).unwrap();
    password.compare(&context, &hash, secret).unwrap();
    assert!(logger.entries.lock().unwrap().is_empty());

    let long_secret = "secret".repeat(13);
    let wrong_secret = "wrong private password";
    let invalid_hash = "private invalid hash";
    let errors = [
        password.hash(&context, &long_secret).unwrap_err(),
        password.compare(&context, &hash, wrong_secret).unwrap_err(),
        password
            .compare(&context, invalid_hash, secret)
            .unwrap_err(),
    ];

    let entries = logger.entries.lock().unwrap();
    assert_eq!(entries.len(), errors.len());
    for (index, (entry, error)) in entries.iter().zip(&errors).enumerate() {
        assert_eq!(entry.context, context);
        assert_eq!(entry.level, LoggerLevel::Error);
        assert_eq!(
            entry.tag,
            if index == 0 {
                "utility/password/bcrypt/hash"
            } else {
                "utility/password/bcrypt/compare"
            }
        );
        assert_eq!(entry.message, error.to_string());
        assert_eq!(entry.meta["error_code"], error.code().into());
        assert_eq!(
            entry.meta["error"],
            error.source().unwrap_or(error).to_string().into()
        );
        let output = format!("{} {:?}", entry.message, entry.meta);
        for credential in [secret, &long_secret, wrong_secret, invalid_hash, &hash] {
            assert!(!output.contains(credential));
        }
    }
}

#[test]
fn hashes_with_fresh_salts_and_distinguishes_password_mismatch() {
    let password: Arc<dyn Password> = Arc::new(new_password(4));
    let first = password.hash(&AppContext::default(), "12345678").unwrap();
    let second = password.hash(&AppContext::default(), "12345678").unwrap();
    assert_ne!(first, second);
    assert_eq!(first.parse::<bcrypt::HashParts>().unwrap().get_cost(), 4);
    password
        .compare(&AppContext::default(), &first, "12345678")
        .unwrap();
    password
        .compare(&AppContext::default(), &second, "12345678")
        .unwrap();

    let error = password
        .compare(&AppContext::default(), &first, "wrong password")
        .unwrap_err();
    assert!(matches!(error, PasswordError::Mismatch));
    assert_eq!(error.code(), "UNAUTHORIZED");
    assert_eq!(error.to_string(), "password does not match");
    assert!(error.source().is_none());
}

#[test]
fn malformed_hash_is_a_failure_with_its_original_cause() {
    let password = new_password(4);
    for hash in ["", "not a bcrypt hash", "$2a$10$short", "é"] {
        let error = password
            .compare(&AppContext::default(), hash, "12345678")
            .unwrap_err();
        assert!(matches!(error, PasswordError::CompareFailure { .. }));
        assert_eq!(error.code(), "FAILURE");
        assert_eq!(error.to_string(), "failed to compare password");
        assert!(
            error
                .source()
                .unwrap()
                .downcast_ref::<bcrypt::BcryptError>()
                .is_some()
        );
    }
}

#[test]
fn hashing_enforces_the_go_limit_in_bytes_including_the_72_byte_boundary() {
    let password = new_password(4);
    for value in ["a".repeat(72), "é".repeat(36)] {
        let hash = password.hash(&AppContext::default(), &value).unwrap();
        password
            .compare(&AppContext::default(), &hash, &value)
            .unwrap();
    }
    for value in ["a".repeat(73), "é".repeat(37)] {
        let error = password.hash(&AppContext::default(), &value).unwrap_err();
        assert!(matches!(error, PasswordError::HashFailure { .. }));
        assert_eq!(error.code(), "FAILURE");
        assert_eq!(error.to_string(), "failed to hash password");
        assert!(matches!(
            error.source().unwrap().downcast_ref::<bcrypt::BcryptError>(),
            Some(bcrypt::BcryptError::Truncation(length)) if *length == value.len()
        ));
    }
}

#[test]
fn invalid_costs_fall_back_to_nadis_default() {
    for cost in [3, 32] {
        let hash = new_password(cost)
            .hash(&AppContext::default(), "12345678")
            .unwrap();
        assert_eq!(hash.parse::<bcrypt::HashParts>().unwrap().get_cost(), 10);
    }
}

#[test]
fn compares_an_existing_go_bcrypt_hash_using_its_stored_cost() {
    // TestBcryptingIsCorrect fixture from golang.org/x/crypto/bcrypt/bcrypt_test.go.
    let hash = "$2a$10$XajjQvNhvvRt5GSeFk1xFeyqRrsxkhBkUiQeg0dt.wU1qD4aFDcga";
    new_password(4)
        .compare(&AppContext::default(), hash, "allmine")
        .unwrap();
}
