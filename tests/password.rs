use std::{error::Error as _, sync::Arc};

use lockmate::{
    domain::{contracts::utility::Password, models::PasswordError},
    infrastructure::utility::password::BcryptPassword,
};

#[test]
fn hashes_with_fresh_salts_and_distinguishes_password_mismatch() {
    let password: Arc<dyn Password> = Arc::new(BcryptPassword::new(4));
    let first = password.hash("12345678").unwrap();
    let second = password.hash("12345678").unwrap();
    assert_ne!(first, second);
    assert_eq!(first.parse::<bcrypt::HashParts>().unwrap().get_cost(), 4);
    password.compare(&first, "12345678").unwrap();
    password.compare(&second, "12345678").unwrap();

    let error = password.compare(&first, "wrong password").unwrap_err();
    assert!(matches!(error, PasswordError::Mismatch));
    assert_eq!(error.code(), "UNAUTHORIZED");
    assert_eq!(error.to_string(), "password does not match");
    assert!(error.source().is_none());
}

#[test]
fn malformed_hash_is_a_failure_with_its_original_cause() {
    let password = BcryptPassword::new(4);
    for hash in ["", "not a bcrypt hash", "$2a$10$short", "é"] {
        let error = password.compare(hash, "12345678").unwrap_err();
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
    let password = BcryptPassword::new(4);
    for value in ["a".repeat(72), "é".repeat(36)] {
        let hash = password.hash(&value).unwrap();
        password.compare(&hash, &value).unwrap();
    }
    for value in ["a".repeat(73), "é".repeat(37)] {
        let error = password.hash(&value).unwrap_err();
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
        let hash = BcryptPassword::new(cost).hash("12345678").unwrap();
        assert_eq!(hash.parse::<bcrypt::HashParts>().unwrap().get_cost(), 10);
    }
}

#[test]
fn compares_an_existing_go_bcrypt_hash_using_its_stored_cost() {
    // TestBcryptingIsCorrect fixture from golang.org/x/crypto/bcrypt/bcrypt_test.go.
    let hash = "$2a$10$XajjQvNhvvRt5GSeFk1xFeyqRrsxkhBkUiQeg0dt.wU1qD4aFDcga";
    BcryptPassword::new(4).compare(hash, "allmine").unwrap();
}
