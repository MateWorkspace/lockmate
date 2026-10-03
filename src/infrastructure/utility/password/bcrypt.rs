use std::{error::Error as _, sync::Arc};

use crate::domain::{
    contracts::utility::{Logger, Password},
    models::{AppContext, LoggerMeta, LoggerMetaValue, PasswordError},
};

const MIN_COST: u32 = 4;
const MAX_COST: u32 = 31;
const DEFAULT_COST: u32 = 10;
const MAX_PASSWORD_BYTES: usize = 72;

pub struct BcryptPassword {
    logger: Arc<dyn Logger>,
    cost: u32,
}

impl BcryptPassword {
    pub fn new(logger: Arc<dyn Logger>, cost: u32) -> Self {
        let cost = if (MIN_COST..=MAX_COST).contains(&cost) {
            cost
        } else {
            DEFAULT_COST
        };
        Self { logger, cost }
    }

    fn log_error(&self, context: &AppContext, tag: &str, error: &PasswordError) {
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

impl Password for BcryptPassword {
    fn hash(&self, context: &AppContext, password: &str) -> Result<String, PasswordError> {
        const TAG: &str = "utility/password/bcrypt/hash";

        let result = if password.len() > MAX_PASSWORD_BYTES {
            Err(PasswordError::HashFailure {
                source: Box::new(bcrypt::BcryptError::Truncation(password.len())),
            })
        } else {
            bcrypt::hash(password, self.cost).map_err(|source| PasswordError::HashFailure {
                source: Box::new(source),
            })
        };
        result.inspect_err(|error| self.log_error(context, TAG, error))
    }

    fn compare(
        &self,
        context: &AppContext,
        stored_hash: &str,
        password: &str,
    ) -> Result<(), PasswordError> {
        const TAG: &str = "utility/password/bcrypt/compare";

        let result = match bcrypt::verify(password, stored_hash) {
            Ok(true) => Ok(()),
            Ok(false) => Err(PasswordError::Mismatch),
            Err(source) => Err(PasswordError::CompareFailure {
                source: Box::new(source),
            }),
        };
        result.inspect_err(|error| self.log_error(context, TAG, error))
    }
}
