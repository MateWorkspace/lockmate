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
        name: "User".into(),
        username: "user".into(),
        role: "admin".into(),
        permissions: vec!["profile.read".into(), "profile.update".into()],
    }
}

fn payload() -> Value {
    let mut payload = serde_json::to_value(access_claims()).unwrap();
    let now = jsonwebtoken::get_current_timestamp();
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
        name: access.name.clone(),
        username: access.username.clone(),
    };
    let encoded_access = token.generate_access(&context, &access).unwrap();
    let encoded_refresh = token.generate_refresh(&context, &refresh).unwrap();
    assert_eq!(
        token.validate_access(&context, &encoded_access).unwrap(),
        access
    );
    assert_eq!(
        token.validate_refresh(&context, &encoded_refresh).unwrap(),
        refresh
    );

    for (encoded, secret, duration) in [
        (&encoded_access, ACCESS_SECRET, 900),
        (&encoded_refresh, REFRESH_SECRET, 86400),
    ] {
        let decoded = decode::<Value>(
            encoded,
            &DecodingKey::from_secret(secret.as_bytes()),
            &Validation::new(Algorithm::HS256),
        )
        .unwrap();
        let claims = decoded.claims;
        assert_eq!(decoded.header.alg, Algorithm::HS256);
        assert_eq!(claims["sub"], access.user_id.to_string());
        assert_eq!(claims["user_id"], access.user_id);
        assert_eq!(claims["iat"], claims["nbf"]);
        assert_eq!(
            claims["exp"].as_i64().unwrap() - claims["iat"].as_i64().unwrap(),
            duration
        );
        assert!(claims.get("claims").is_none());
    }
    assert!(matches!(
        token.validate_access(&context, &encoded_refresh),
        Err(TokenError::Invalid)
    ));
    assert!(matches!(
        token.validate_refresh(&context, &encoded_access),
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
            token.validate_access(&context, &value),
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
            token.validate_access(&context, &encoded),
            Err(TokenError::Expired)
        ));
        let encoded = sign(&claims, REFRESH_SECRET, Algorithm::HS256);
        assert!(matches!(
            token.validate_refresh(&context, &encoded),
            Err(TokenError::Expired)
        ));
    }
    let mut claims = payload();
    claims["nbf"] = (jsonwebtoken::get_current_timestamp() + 3600).into();
    assert!(matches!(
        token.validate_access(&context, &sign(&claims, ACCESS_SECRET, Algorithm::HS256)),
        Err(TokenError::Invalid)
    ));
    claims["nbf"] = json!("invalid time");
    assert!(matches!(
        token.validate_access(&context, &sign(&claims, ACCESS_SECRET, Algorithm::HS256)),
        Err(TokenError::Invalid)
    ));
}

#[test]
fn supports_nadi_hmac_algorithms_and_subject_fallback() {
    let token = token();
    let context = AppContext::default();
    for algorithm in [Algorithm::HS256, Algorithm::HS384, Algorithm::HS512] {
        let mut claims = payload();
        claims.as_object_mut().unwrap().remove("user_id");
        let encoded = sign(&claims, ACCESS_SECRET, algorithm);
        assert_eq!(
            token.validate_access(&context, &encoded).unwrap(),
            access_claims()
        );
        claims["user_id"] = 0.into();
        assert_eq!(
            token
                .validate_access(&context, &sign(&claims, ACCESS_SECRET, algorithm))
                .unwrap(),
            access_claims()
        );
        claims["sub"] = "not an id".into();
        assert!(matches!(
            token.validate_access(&context, &sign(&claims, ACCESS_SECRET, algorithm)),
            Err(TokenError::Invalid)
        ));
        // A nonzero user_id takes precedence over sub, as in Nadi.
        claims["user_id"] = access_claims().user_id.into();
        assert_eq!(
            token
                .validate_access(&context, &sign(&claims, ACCESS_SECRET, algorithm))
                .unwrap(),
            access_claims()
        );
    }
    let mut claims = payload();
    for field in ["exp", "iat", "nbf", "sub"] {
        claims.as_object_mut().unwrap().remove(field);
    }
    assert_eq!(
        token
            .validate_access(&context, &sign(&claims, ACCESS_SECRET, Algorithm::HS256))
            .unwrap(),
        access_claims()
    );
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
    };
    let refresh = RefreshClaims {
        user_id: 1,
        name: "User".into(),
        username: "user".into(),
    };
    let errors = [
        token
            .generate_access(&context, &access_claims())
            .unwrap_err(),
        token.generate_refresh(&context, &refresh).unwrap_err(),
        token
            .validate_access(&context, "private access token")
            .unwrap_err(),
        token
            .validate_refresh(&context, "private refresh token")
            .unwrap_err(),
        token.validate_access(&context, "").unwrap_err(),
        token.validate_refresh(&context, "").unwrap_err(),
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
        name: "User".into(),
        username: "user".into(),
    };
    assert!(matches!(
        token.generate_refresh(&context, &refresh),
        Err(TokenError::Failure { .. })
    ));
}
