use std::{
    error::Error as _,
    sync::{Arc, Mutex},
    time::Duration,
};

use jsonwebtoken::{Algorithm, DecodingKey, EncodingKey, Header, Validation, decode, encode};
use lockmate::{
    domain::{
        contracts::utility::{Logger, Token},
        models::{AccessClaims, AppContext, LoggerLevel, LoggerMeta, RefreshClaims, TokenError},
    },
    infrastructure::utility::{logger::JsonLogger, token::JwtToken},
};
use serde_json::{Value, json};

const SPACE_ID: i64 = 7;
const MEMBER_ID: i64 = 42;

const ACCESS_SECRET: &str = "access-secret-for-tests-only-0123456789abcdefghijklmnopqrstuvwxyz-XYZ";
const REFRESH_SECRET: &str =
    "refresh-secret-for-tests-only-0123456789abcdefghijklmnopqrstuvwxyz-XYZ";

fn token() -> JwtToken {
    JwtToken::new(
        Arc::new(JsonLogger::with_writer(std::io::sink(), LoggerLevel::Error)),
        ACCESS_SECRET,
        REFRESH_SECRET,
        Duration::from_secs(900),
        Duration::from_secs(86400),
    )
}

fn access_claims() -> AccessClaims {
    AccessClaims {
        user_id: 9_007_199_254_740_993,
        space_id: SPACE_ID,
        member_id: MEMBER_ID,
        name: "User".into(),
        username: "user".into(),
        roles: vec!["admin".into(), "user".into()],
        permissions: vec!["profile.read".into(), "profile.update".into()],
    }
}

fn payload() -> Value {
    let mut payload = serde_json::to_value(access_claims()).unwrap();
    let now = jsonwebtoken::get_current_timestamp();
    payload["aud"] = "lockmate:space:7".into();
    payload["token_type"] = "access".into();
    payload["sub"] = access_claims().user_id.to_string().into();
    payload["iat"] = now.into();
    payload["nbf"] = now.into();
    payload["exp"] = (now + 900).into();
    payload
}

fn sign(payload: &Value, secret: &str, algorithm: Algorithm) -> String {
    encode(
        &Header::new(algorithm),
        payload,
        &EncodingKey::from_secret(secret.as_bytes()),
    )
    .unwrap()
}

#[test]
fn round_trips_both_claim_types_with_independent_lifetimes_and_registered_claims() {
    let token: Arc<dyn Token> = Arc::new(token());
    let context = AppContext::default();
    let access = access_claims();
    let refresh = RefreshClaims {
        user_id: access.user_id,
        space_id: SPACE_ID,
        member_id: MEMBER_ID,
        name: access.name.clone(),
        username: access.username.clone(),
    };
    let encoded_access = token.generate_access(&context, &access).unwrap();
    let encoded_refresh = token.generate_refresh(&context, &refresh).unwrap();
    assert_eq!(
        token
            .validate_access(&context, SPACE_ID, &encoded_access)
            .unwrap(),
        access
    );
    assert_eq!(
        token
            .validate_refresh(&context, SPACE_ID, &encoded_refresh)
            .unwrap(),
        refresh
    );

    for (encoded, secret, duration) in [
        (&encoded_access, ACCESS_SECRET, 900),
        (&encoded_refresh, REFRESH_SECRET, 86400),
    ] {
        let mut validation = Validation::new(Algorithm::HS256);
        validation.set_audience(&["lockmate:space:7"]);
        let decoded = decode::<Value>(
            encoded,
            &DecodingKey::from_secret(secret.as_bytes()),
            &validation,
        )
        .unwrap();
        let claims = decoded.claims;
        assert_eq!(decoded.header.alg, Algorithm::HS256);
        assert_eq!(claims["sub"], access.user_id.to_string());
        assert_eq!(claims["user_id"], access.user_id);
        assert_eq!(claims["space_id"], SPACE_ID);
        assert_eq!(claims["member_id"], MEMBER_ID);
        assert_eq!(claims["aud"], "lockmate:space:7");
        assert_eq!(
            claims["token_type"],
            if duration == 900 { "access" } else { "refresh" }
        );
        assert_eq!(claims["iat"], claims["nbf"]);
        assert_eq!(
            claims["exp"].as_i64().unwrap() - claims["iat"].as_i64().unwrap(),
            duration
        );
        assert!(claims.get("claims").is_none());
        if duration == 86400 {
            assert!(claims.get("roles").is_none());
            assert!(claims.get("permissions").is_none());
        }
    }
    assert!(matches!(
        token.validate_access(&context, SPACE_ID, &encoded_refresh),
        Err(TokenError::Invalid)
    ));
    assert!(matches!(
        token.validate_refresh(&context, SPACE_ID, &encoded_access),
        Err(TokenError::Invalid)
    ));
}

#[test]
fn rejects_bad_signatures_tampering_malformed_tokens_and_non_hmac_headers() {
    let token = token();
    let context = AppContext::default();
    let good = sign(&payload(), ACCESS_SECRET, Algorithm::HS256);
    let pieces: Vec<_> = good.split('.').collect();
    let mut signature = pieces[2].as_bytes().to_vec();
    signature[0] = if signature[0] == b'A' { b'B' } else { b'A' };
    let tampered = format!(
        "{}.{}.{}",
        pieces[0],
        pieces[1],
        String::from_utf8(signature).unwrap()
    );
    let unsigned = format!("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.{}.", pieces[1]);
    let rsa = format!(
        "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.{}.{}",
        pieces[1], pieces[2]
    );
    for value in [
        "malformed".into(),
        tampered,
        unsigned,
        rsa,
        sign(&payload(), REFRESH_SECRET, Algorithm::HS256),
    ] {
        assert!(matches!(
            token.validate_access(&context, SPACE_ID, &value),
            Err(TokenError::Invalid)
        ));
    }
}

#[test]
fn validates_expiration_and_not_before_without_clock_skew() {
    let token = token();
    let context = AppContext::default();
    for expired_at in [1, jsonwebtoken::get_current_timestamp()] {
        let mut claims = payload();
        claims["exp"] = expired_at.into();
        let encoded = sign(&claims, ACCESS_SECRET, Algorithm::HS256);
        assert!(matches!(
            token.validate_access(&context, SPACE_ID, &encoded),
            Err(TokenError::Expired)
        ));
        let encoded = sign(&claims, REFRESH_SECRET, Algorithm::HS256);
        assert!(matches!(
            token.validate_refresh(&context, SPACE_ID, &encoded),
            Err(TokenError::Expired)
        ));
    }
    let mut claims = payload();
    claims["nbf"] = (jsonwebtoken::get_current_timestamp() + 3600).into();
    assert!(matches!(
        token.validate_access(
            &context,
            SPACE_ID,
            &sign(&claims, ACCESS_SECRET, Algorithm::HS256)
        ),
        Err(TokenError::Invalid)
    ));
    claims["nbf"] = json!("invalid time");
    assert!(matches!(
        token.validate_access(
            &context,
            SPACE_ID,
            &sign(&claims, ACCESS_SECRET, Algorithm::HS256)
        ),
        Err(TokenError::Invalid)
    ));
}

#[test]
fn supports_existing_hmac_algorithms_with_strict_subject_consistency() {
    let token = token();
    let context = AppContext::default();
    for algorithm in [Algorithm::HS256, Algorithm::HS384, Algorithm::HS512] {
        let claims = payload();
        assert_eq!(
            token
                .validate_access(&context, SPACE_ID, &sign(&claims, ACCESS_SECRET, algorithm))
                .unwrap(),
            access_claims()
        );
        for subject in ["not an id", "1", "09007199254740993"] {
            let mut claims = payload();
            claims["sub"] = subject.into();
            assert!(matches!(
                token.validate_access(&context, SPACE_ID, &sign(&claims, ACCESS_SECRET, algorithm)),
                Err(TokenError::Invalid)
            ));
        }
    }
}

#[test]
fn requires_typed_identity_scope_purpose_and_registered_fields() {
    let token = token();
    let context = AppContext::default();
    for field in [
        "user_id",
        "space_id",
        "member_id",
        "name",
        "username",
        "roles",
        "permissions",
        "exp",
        "iat",
        "nbf",
        "sub",
        "aud",
        "token_type",
    ] {
        let mut claims = payload();
        claims.as_object_mut().unwrap().remove(field);
        assert!(
            matches!(
                token.validate_access(
                    &context,
                    SPACE_ID,
                    &sign(&claims, ACCESS_SECRET, Algorithm::HS256)
                ),
                Err(TokenError::Invalid)
            ),
            "missing {field}"
        );
        claims[field] = Value::Null;
        assert!(
            matches!(
                token.validate_access(
                    &context,
                    SPACE_ID,
                    &sign(&claims, ACCESS_SECRET, Algorithm::HS256)
                ),
                Err(TokenError::Invalid)
            ),
            "null {field}"
        );
    }
    for field in ["user_id", "space_id", "member_id"] {
        for value in [json!(0), json!(-1), json!(1.5), json!("7"), json!(true)] {
            let mut claims = payload();
            claims[field] = value;
            assert!(
                matches!(
                    token.validate_access(
                        &context,
                        SPACE_ID,
                        &sign(&claims, ACCESS_SECRET, Algorithm::HS256)
                    ),
                    Err(TokenError::Invalid)
                ),
                "invalid {field}"
            );
        }
    }
    for (field, value) in [
        ("aud", json!(["lockmate:space:7"])),
        ("exp", json!("9999999999")),
        ("iat", json!(1.5)),
        ("nbf", json!(true)),
        ("roles", json!("admin")),
        ("permissions", json!([7])),
        ("sub", json!(9007199254740993i64)),
    ] {
        let mut claims = payload();
        claims[field] = value;
        assert!(
            matches!(
                token.validate_access(
                    &context,
                    SPACE_ID,
                    &sign(&claims, ACCESS_SECRET, Algorithm::HS256)
                ),
                Err(TokenError::Invalid)
            ),
            "invalid {field}"
        );
    }
    let legacy = json!({"user_id":1,"name":"User","username":"user","role":"admin","permissions":[],"sub":"1","exp":jsonwebtoken::get_current_timestamp()+900});
    assert!(matches!(
        token.validate_access(
            &context,
            SPACE_ID,
            &sign(&legacy, ACCESS_SECRET, Algorithm::HS256)
        ),
        Err(TokenError::Invalid)
    ));
}

#[test]
fn rejects_cross_space_audience_and_token_type_substitution_with_identical_secrets() {
    let context = AppContext::default();
    let token = JwtToken::new(
        Arc::new(JsonLogger::with_writer(std::io::sink(), LoggerLevel::None)),
        ACCESS_SECRET,
        ACCESS_SECRET,
        Duration::from_secs(900),
        Duration::from_secs(86400),
    );
    let access = access_claims();
    let refresh = RefreshClaims {
        user_id: access.user_id,
        space_id: SPACE_ID,
        member_id: MEMBER_ID,
        name: access.name.clone(),
        username: access.username.clone(),
    };
    let encoded_access = token.generate_access(&context, &access).unwrap();
    let encoded_refresh = token.generate_refresh(&context, &refresh).unwrap();
    assert!(matches!(
        token.validate_access(&context, SPACE_ID, &encoded_refresh),
        Err(TokenError::Invalid)
    ));
    assert!(matches!(
        token.validate_refresh(&context, SPACE_ID, &encoded_access),
        Err(TokenError::Invalid)
    ));
    assert!(matches!(
        token.validate_access(&context, SPACE_ID + 1, &encoded_access),
        Err(TokenError::Invalid)
    ));
    assert!(matches!(
        token.validate_refresh(&context, SPACE_ID + 1, &encoded_refresh),
        Err(TokenError::Invalid)
    ));
    for (field, value) in [
        ("space_id", json!(8)),
        ("aud", json!("lockmate:space:8")),
        ("token_type", json!("refresh")),
    ] {
        let mut claims = payload();
        claims[field] = value;
        assert!(matches!(
            token.validate_access(
                &context,
                SPACE_ID,
                &sign(&claims, ACCESS_SECRET, Algorithm::HS256)
            ),
            Err(TokenError::Invalid)
        ));
    }
}

#[test]
fn validates_issue_time_and_ordered_time_claims() {
    let token = token();
    let context = AppContext::default();
    let now = jsonwebtoken::get_current_timestamp();
    for (iat, nbf, exp) in [
        (now + 10, now, now + 900),
        (now, now - 10, now + 900),
        (now, now + 900, now + 900),
    ] {
        let mut claims = payload();
        claims["iat"] = iat.into();
        claims["nbf"] = nbf.into();
        claims["exp"] = exp.into();
        assert!(matches!(
            token.validate_access(
                &context,
                SPACE_ID,
                &sign(&claims, ACCESS_SECRET, Algorithm::HS256)
            ),
            Err(TokenError::Invalid)
        ));
    }
}

#[test]
fn empty_grants_and_large_identity_values_round_trip() {
    let context = AppContext::default();
    let token = token();
    let mut claims = access_claims();
    claims.space_id = i64::MAX;
    claims.member_id = i64::MAX - 1;
    claims.roles.clear();
    claims.permissions.clear();
    let encoded = token.generate_access(&context, &claims).unwrap();
    assert_eq!(
        token.validate_access(&context, i64::MAX, &encoded).unwrap(),
        claims
    );
}

#[test]
fn invalid_generation_ids_expected_spaces_and_zero_lifetimes_are_classified() {
    let context = AppContext::default();
    let token = token();
    for field in ["user_id", "space_id", "member_id"] {
        for id in [0, -1] {
            let mut claims = access_claims();
            match field {
                "user_id" => claims.user_id = id,
                "space_id" => claims.space_id = id,
                _ => claims.member_id = id,
            }
            assert!(matches!(
                token.generate_access(&context, &claims),
                Err(TokenError::BadArgs)
            ));
            let refresh = RefreshClaims {
                user_id: claims.user_id,
                space_id: claims.space_id,
                member_id: claims.member_id,
                name: claims.name,
                username: claims.username,
            };
            assert!(matches!(
                token.generate_refresh(&context, &refresh),
                Err(TokenError::BadArgs)
            ));
        }
    }
    for space_id in [0, -1] {
        assert!(matches!(
            token.validate_access(&context, space_id, "token"),
            Err(TokenError::BadArgs)
        ));
        assert!(matches!(
            token.validate_refresh(&context, space_id, "token"),
            Err(TokenError::BadArgs)
        ));
    }
    let token = JwtToken::new(
        Arc::new(JsonLogger::with_writer(std::io::sink(), LoggerLevel::None)),
        ACCESS_SECRET,
        REFRESH_SECRET,
        Duration::ZERO,
        Duration::ZERO,
    );
    let access = access_claims();
    let refresh = RefreshClaims {
        user_id: access.user_id,
        space_id: access.space_id,
        member_id: access.member_id,
        name: access.name.clone(),
        username: access.username.clone(),
    };
    assert!(matches!(
        token.generate_access(&context, &access),
        Err(TokenError::Failure { .. })
    ));
    assert!(matches!(
        token.generate_refresh(&context, &refresh),
        Err(TokenError::Failure { .. })
    ));
}

struct LogEntry {
    context: AppContext,
    level: LoggerLevel,
    tag: String,
    message: String,
    meta: LoggerMeta,
}

#[derive(Default)]
struct RecordingLogger(Mutex<Vec<LogEntry>>);

impl Logger for RecordingLogger {
    fn log(
        &self,
        context: &AppContext,
        level: LoggerLevel,
        tag: &str,
        message: &str,
        meta: &LoggerMeta,
    ) {
        self.0.lock().unwrap().push(LogEntry {
            context: context.clone(),
            level,
            tag: tag.into(),
            message: message.into(),
            meta: meta.clone(),
        });
    }
}

#[test]
fn configuration_and_input_errors_are_logged_once_without_secrets_or_tokens() {
    let logger = Arc::new(RecordingLogger::default());
    let token = JwtToken::new(
        logger.clone(),
        "",
        "",
        Duration::from_secs(900),
        Duration::from_secs(86400),
    );
    let context = AppContext {
        actor: Some("admin".into()),
        trace_id: Some(uuid::Uuid::from_u128(42)),
        ..AppContext::default()
    };
    let refresh = RefreshClaims {
        user_id: 1,
        space_id: SPACE_ID,
        member_id: MEMBER_ID,
        name: "User".into(),
        username: "user".into(),
    };
    let errors = [
        token
            .generate_access(&context, &access_claims())
            .unwrap_err(),
        token.generate_refresh(&context, &refresh).unwrap_err(),
        token
            .validate_access(&context, SPACE_ID, "private access token")
            .unwrap_err(),
        token
            .validate_refresh(&context, SPACE_ID, "private refresh token")
            .unwrap_err(),
        token.validate_access(&context, SPACE_ID, "").unwrap_err(),
        token.validate_refresh(&context, SPACE_ID, "").unwrap_err(),
    ];
    for error in &errors[..4] {
        assert!(matches!(error, TokenError::Failure { .. }));
        assert!(error.source().is_some());
    }
    for error in &errors[4..] {
        assert!(matches!(error, TokenError::BadArgs));
    }
    let entries = logger.0.lock().unwrap();
    assert_eq!(entries.len(), errors.len());
    for (entry, error) in entries.iter().zip(&errors) {
        assert_eq!(entry.context, context);
        assert_eq!(entry.level, LoggerLevel::Error);
        assert!(entry.tag.starts_with("utility/token/jwt/"));
        assert_eq!(entry.message, error.to_string());
        assert_eq!(entry.meta["error_code"], error.code().into());
        let output = format!("{} {:?}", entry.message, entry.meta);
        for secret in [
            "private access token",
            "private refresh token",
            ACCESS_SECRET,
            REFRESH_SECRET,
        ] {
            assert!(!output.contains(secret));
        }
    }
}

#[test]
fn unrepresentable_lifetimes_return_failure_instead_of_panicking() {
    let token = JwtToken::new(
        Arc::new(JsonLogger::with_writer(std::io::sink(), LoggerLevel::Error)),
        ACCESS_SECRET,
        REFRESH_SECRET,
        Duration::from_secs(u64::MAX),
        Duration::from_secs(u64::MAX),
    );
    let context = AppContext::default();
    assert!(matches!(
        token.generate_access(&context, &access_claims()),
        Err(TokenError::Failure { .. })
    ));
    let refresh = RefreshClaims {
        user_id: 1,
        space_id: SPACE_ID,
        member_id: MEMBER_ID,
        name: "User".into(),
        username: "user".into(),
    };
    assert!(matches!(
        token.generate_refresh(&context, &refresh),
        Err(TokenError::Failure { .. })
    ));
}

#[test]
fn refresh_validation_requires_its_identity_and_registered_claims_without_grants() {
    let token = token();
    let context = AppContext::default();
    let mut claims = payload();
    claims["token_type"] = "refresh".into();
    claims.as_object_mut().unwrap().remove("roles");
    claims.as_object_mut().unwrap().remove("permissions");
    let expected = RefreshClaims {
        user_id: access_claims().user_id,
        space_id: SPACE_ID,
        member_id: MEMBER_ID,
        name: "User".into(),
        username: "user".into(),
    };
    assert_eq!(
        token
            .validate_refresh(
                &context,
                SPACE_ID,
                &sign(&claims, REFRESH_SECRET, Algorithm::HS256)
            )
            .unwrap(),
        expected
    );
    for field in [
        "user_id",
        "space_id",
        "member_id",
        "name",
        "username",
        "sub",
        "aud",
        "token_type",
        "iat",
        "nbf",
        "exp",
    ] {
        let mut missing = claims.clone();
        missing.as_object_mut().unwrap().remove(field);
        assert!(
            matches!(
                token.validate_refresh(
                    &context,
                    SPACE_ID,
                    &sign(&missing, REFRESH_SECRET, Algorithm::HS256)
                ),
                Err(TokenError::Invalid)
            ),
            "missing refresh {field}"
        );
    }
    let mut invalid = claims.clone();
    invalid["sub"] = "1".into();
    assert!(matches!(
        token.validate_refresh(
            &context,
            SPACE_ID,
            &sign(&invalid, REFRESH_SECRET, Algorithm::HS256)
        ),
        Err(TokenError::Invalid)
    ));
}
