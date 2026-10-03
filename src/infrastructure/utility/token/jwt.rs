use std::{error::Error as _, io, sync::Arc, time::Duration};

use jsonwebtoken::{
    Algorithm, DecodingKey, EncodingKey, Header, Validation, decode, encode, errors::ErrorKind,
};
use serde::{Serialize, de::DeserializeOwned};
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
        user_id: i64,
        secret: &[u8],
        duration: Duration,
    ) -> Result<String, TokenError> {
        if secret.is_empty() {
            return Err(failure("token secret is not configured"));
        }
        let now = OffsetDateTime::now_utc();
        let duration = time::Duration::try_from(duration)
            .map_err(|_| failure("token lifetime is out of range"))?;
        let expires_at = now
            .checked_add(duration)
            .ok_or_else(|| failure("token expiration is out of range"))?;
        let claims = JwtClaims {
            claims,
            sub: user_id.to_string(),
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

    fn validate<T: DeserializeOwned>(&self, token: &str, secret: &[u8]) -> Result<T, TokenError> {
        if token.is_empty() {
            return Err(TokenError::BadArgs);
        }
        if secret.is_empty() {
            return Err(failure("token secret is not configured"));
        }
        let mut validation = Validation::new(Algorithm::HS256);

        validation.algorithms = vec![Algorithm::HS256, Algorithm::HS384, Algorithm::HS512];
        validation.required_spec_claims.clear();
        validation.leeway = 0;
        validation.validate_nbf = true;
        validation.validate_aud = false;

        let mut claims = decode::<Value>(token, &DecodingKey::from_secret(secret), &validation)
            .map_err(|error| match error.kind() {
                ErrorKind::ExpiredSignature => TokenError::Expired,
                _ => TokenError::Invalid,
            })?
            .claims;

        if !claims.is_object() {
            return Err(TokenError::Invalid);
        }

        if claims
            .get("exp")
            .and_then(Value::as_i64)
            .is_some_and(|expiration| expiration <= OffsetDateTime::now_utc().unix_timestamp())
        {
            return Err(TokenError::Expired);
        }

        let user_id = match claims.get("user_id") {
            None | Some(Value::Null) => 0,
            Some(value) => value.as_i64().ok_or(TokenError::Invalid)?,
        };
        let user_id = if user_id == 0 {
            claims
                .get("sub")
                .and_then(Value::as_str)
                .ok_or(TokenError::Invalid)?
                .parse::<i64>()
                .map_err(|_| TokenError::Invalid)?
        } else {
            user_id
        };

        claims["user_id"] = user_id.into();
        serde_json::from_value(claims).map_err(|_| TokenError::Invalid)
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
            claims.user_id,
            &self.access_secret,
            self.access_duration,
        )
        .inspect_err(|error| self.log_error(context, TAG, error))
    }

    fn validate_access(
        &self,
        context: &AppContext,
        token: &str,
    ) -> Result<AccessClaims, TokenError> {
        const TAG: &str = "utility/token/jwt/validate_access";
        self.validate(token, &self.access_secret)
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
            claims.user_id,
            &self.refresh_secret,
            self.refresh_duration,
        )
        .inspect_err(|error| self.log_error(context, TAG, error))
    }

    fn validate_refresh(
        &self,
        context: &AppContext,
        token: &str,
    ) -> Result<RefreshClaims, TokenError> {
        const TAG: &str = "utility/token/jwt/validate_refresh";
        self.validate(token, &self.refresh_secret)
            .inspect_err(|error| self.log_error(context, TAG, error))
    }
}

#[derive(Serialize)]
struct JwtClaims<'a, T> {
    #[serde(flatten)]
    claims: &'a T,
    sub: String,
    iat: i64,
    nbf: i64,
    exp: i64,
}

fn failure(message: &'static str) -> TokenError {
    TokenError::Failure {
        source: Box::new(io::Error::other(message)),
    }
}
