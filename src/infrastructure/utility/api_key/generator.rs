use std::{error::Error as _, sync::Arc};

use sha2::{Digest, Sha256};

use crate::domain::{
    contracts::utility::{ApiKey, Logger},
    models::{ApiKeyError, AppContext, GeneratedApiKey, LoggerMeta, LoggerMetaValue},
};

const KEY_PREFIX: &str = "lockmate-";
const RANDOM_BYTES: usize = 32;
const KEY_LENGTH: usize = KEY_PREFIX.len() + RANDOM_BYTES * 2;
const REDACTED_SIZE: usize = 4;

pub struct ApiKeyGenerator {
    logger: Arc<dyn Logger>,
}

impl ApiKeyGenerator {
    pub fn new(logger: Arc<dyn Logger>) -> Self {
        Self { logger }
    }

    fn generate_with(
        &self,
        context: &AppContext,
        fill: impl FnOnce(&mut [u8]) -> Result<(), getrandom::Error>,
    ) -> Result<GeneratedApiKey, ApiKeyError> {
        const TAG: &str = "utility/api_key/generator/generate";

        let mut bytes = [0; RANDOM_BYTES];
        fill(&mut bytes)
            .map_err(|source| ApiKeyError::Failure {
                source: Box::new(source),
            })
            .map(|()| {
                let raw = format!("{KEY_PREFIX}{}", hex::encode(bytes));
                let hash = self.hash(context, &raw);
                let redacted = raw[KEY_LENGTH - REDACTED_SIZE..].to_owned();
                GeneratedApiKey {
                    raw,
                    hash,
                    redacted,
                }
            })
            .inspect_err(|error| self.log_error(context, TAG, error))
    }

    fn log_error(&self, context: &AppContext, tag: &str, error: &ApiKeyError) {
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

impl ApiKey for ApiKeyGenerator {
    fn generate(&self, context: &AppContext) -> Result<GeneratedApiKey, ApiKeyError> {
        self.generate_with(context, getrandom::fill)
    }

    fn hash(&self, _context: &AppContext, raw: &str) -> String {
        hex::encode(Sha256::digest(raw.as_bytes()))
    }

    fn validate(&self, context: &AppContext, raw: &str) -> Result<(), ApiKeyError> {
        const TAG: &str = "utility/api_key/generator/validate";

        let result = if raw.is_empty() {
            Err(ApiKeyError::BadArgs)
        } else if raw.len() != KEY_LENGTH
            || !raw.starts_with(KEY_PREFIX)
            || !raw.as_bytes()[KEY_PREFIX.len()..]
                .iter()
                .all(|byte| byte.is_ascii_digit() || (b'a'..=b'f').contains(byte))
        {
            Err(ApiKeyError::Invalid)
        } else {
            Ok(())
        };
        result.inspect_err(|error| self.log_error(context, TAG, error))
    }
}
