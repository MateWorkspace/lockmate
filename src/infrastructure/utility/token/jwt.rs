use std::{error::Error as _, io, sync::Arc, time::Duration};

use jsonwebtoken::{
    Algorithm, DecodingKey, EncodingKey, Header, Validation, decode, encode, errors::ErrorKind,
};
use serde::{Deserialize, Serialize, de::DeserializeOwned};
use serde_json::Value;
use time::OffsetDateTime;

use crate::domain::{
    contracts::utility::{Logger, Token},
    models::{AccessClaims, AppContext, LoggerMeta, LoggerMetaValue, RefreshClaims, TokenError},
};

pub struct JwtToken {
    logger: Arc<dyn Logger>,
    access_secret: Vec<u8>,
    refresh_secret: Vec<u8>,
    access_duration: Duration,
    refresh_duration: Duration,
}

impl JwtToken {
    pub fn new(
        logger: Arc<dyn Logger>,
        access_secret: &str,
        refresh_secret: &str,
        access_duration: Duration,
        refresh_duration: Duration,
    ) -> Self {
        Self {
            logger,
            access_secret: access_secret.as_bytes().to_vec(),
            refresh_secret: refresh_secret.as_bytes().to_vec(),
            access_duration,
            refresh_duration,
        }
    }

    fn generate<T: Serialize>(
        &self,
        claims: &T,
        identity: (i64, i64, i64),
        token_type: &str,
        secret: &[u8],
        duration: Duration,
    ) -> Result<String, TokenError> {
        let (user_id, space_id, member_id) = identity;
        if user_id <= 0 || space_id <= 0 || member_id <= 0 {
            return Err(TokenError::BadArgs);
        }
        if secret.is_empty() {
            return Err(failure("token secret is not configured"));
        }
        if duration.is_zero() {
            return Err(failure("token lifetime must be positive"));
        }
        let now = OffsetDateTime::now_utc();
        let duration = time::Duration::try_from(duration)
            .map_err(|_| failure("token lifetime is out of range"))?;
        let expires_at = now
            .checked_add(duration)
            .ok_or_else(|| failure("token expiration is out of range"))?;
        if expires_at.unix_timestamp() <= now.unix_timestamp() {
            return Err(failure("token lifetime is below timestamp resolution"));
        }
        let claims = JwtClaims {
            claims,
            sub: user_id.to_string(),
            aud: audience(space_id),
            token_type: token_type.to_owned(),
            iat: now.unix_timestamp(),
            nbf: now.unix_timestamp(),
            exp: expires_at.unix_timestamp(),
        };
        encode(
            &Header::new(Algorithm::HS256),
            &claims,
            &EncodingKey::from_secret(secret),
        )
        .map_err(|source| TokenError::Failure {
            source: Box::new(source),
        })
    }

    fn validate<T: DeserializeOwned>(
        &self,
        space_id: i64,
        token: &str,
        token_type: &str,
        secret: &[u8],
    ) -> Result<T, TokenError> {
        if token.is_empty() || space_id <= 0 {
            return Err(TokenError::BadArgs);
        }
        if secret.is_empty() {
            return Err(failure("token secret is not configured"));
        }
        let expected_audience = audience(space_id);
        let mut validation = Validation::new(Algorithm::HS256);
        validation.algorithms = vec![Algorithm::HS256, Algorithm::HS384, Algorithm::HS512];
        validation.set_required_spec_claims(&["sub", "aud", "exp", "nbf"]);
        validation.set_audience(&[&expected_audience]);
        validation.leeway = 0;
        validation.validate_nbf = true;
        // Typed decoding also requires iat and rejects malformed registered fields.
        let claims =
            decode::<JwtClaims<Value>>(token, &DecodingKey::from_secret(secret), &validation)
                .map_err(|error| match error.kind() {
                    ErrorKind::ExpiredSignature => TokenError::Expired,
                    _ => TokenError::Invalid,
                })?
                .claims;
        let user_id = positive_id(&claims.claims, "user_id")?;
        let claimed_space = positive_id(&claims.claims, "space_id")?;
        positive_id(&claims.claims, "member_id")?;
        let now = OffsetDateTime::now_utc().unix_timestamp();
        if claims.exp <= now {
            return Err(TokenError::Expired);
        }
        if claimed_space != space_id
            || claims.aud != expected_audience
            || claims.sub != user_id.to_string()
            || claims.token_type != token_type
            || claims.iat > now
            || claims.nbf > now
            || claims.iat > claims.nbf
            || claims.nbf >= claims.exp
        {
            return Err(TokenError::Invalid);
        }
        serde_json::from_value(claims.claims).map_err(|_| TokenError::Invalid)
    }

    fn log_error(&self, context: &AppContext, tag: &str, error: &TokenError) {
        let meta = LoggerMeta::from([
            ("error_code".into(), error.code().into()),
            (
                "error".into(),
                LoggerMetaValue::from_error(error.source().unwrap_or(error)),
            ),
        ]);
        self.logger.error(context, tag, &error.to_string(), &meta);
    }
}

impl Token for JwtToken {
    fn generate_access(
        &self,
        context: &AppContext,
        claims: &AccessClaims,
    ) -> Result<String, TokenError> {
        const TAG: &str = "utility/token/jwt/generate_access";
        self.generate(
            claims,
            (claims.user_id, claims.space_id, claims.member_id),
            "access",
            &self.access_secret,
            self.access_duration,
        )
        .inspect_err(|error| self.log_error(context, TAG, error))
    }

    fn validate_access(
        &self,
        context: &AppContext,
        space_id: i64,
        token: &str,
    ) -> Result<AccessClaims, TokenError> {
        const TAG: &str = "utility/token/jwt/validate_access";
        self.validate(space_id, token, "access", &self.access_secret)
            .inspect_err(|error| self.log_error(context, TAG, error))
    }

    fn generate_refresh(
        &self,
        context: &AppContext,
        claims: &RefreshClaims,
    ) -> Result<String, TokenError> {
        const TAG: &str = "utility/token/jwt/generate_refresh";
        self.generate(
            claims,
            (claims.user_id, claims.space_id, claims.member_id),
            "refresh",
            &self.refresh_secret,
            self.refresh_duration,
        )
        .inspect_err(|error| self.log_error(context, TAG, error))
    }

    fn validate_refresh(
        &self,
        context: &AppContext,
        space_id: i64,
        token: &str,
    ) -> Result<RefreshClaims, TokenError> {
        const TAG: &str = "utility/token/jwt/validate_refresh";
        self.validate(space_id, token, "refresh", &self.refresh_secret)
            .inspect_err(|error| self.log_error(context, TAG, error))
    }
}

#[derive(Serialize, Deserialize)]
struct JwtClaims<T> {
    #[serde(flatten)]
    claims: T,
    sub: String,
    aud: String,
    token_type: String,
    iat: i64,
    nbf: i64,
    exp: i64,
}

fn failure(message: &'static str) -> TokenError {
    TokenError::Failure {
        source: Box::new(io::Error::other(message)),
    }
}

fn audience(space_id: i64) -> String {
    format!("lockmate:space:{space_id}")
}

fn positive_id(claims: &Value, field: &str) -> Result<i64, TokenError> {
    claims
        .get(field)
        .and_then(Value::as_i64)
        .filter(|id| *id > 0)
        .ok_or(TokenError::Invalid)
}
