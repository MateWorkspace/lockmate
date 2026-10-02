use crate::domain::{contracts::utility::Password, models::PasswordError};

const MIN_COST: u32 = 4;
const MAX_COST: u32 = 31;
const DEFAULT_COST: u32 = 10;
const MAX_PASSWORD_BYTES: usize = 72;

pub struct BcryptPassword {
    cost: u32,
}

impl BcryptPassword {
    pub fn new(cost: u32) -> Self {
        let cost = if (MIN_COST..=MAX_COST).contains(&cost) {
            cost
        } else {
            DEFAULT_COST
        };
        Self { cost }
    }
}

impl Password for BcryptPassword {
    fn hash(&self, password: &str) -> Result<String, PasswordError> {
        if password.len() > MAX_PASSWORD_BYTES {
            return Err(PasswordError::HashFailure {
                source: Box::new(bcrypt::BcryptError::Truncation(password.len())),
            });
        }

        bcrypt::hash(password, self.cost).map_err(|source| PasswordError::HashFailure {
            source: Box::new(source),
        })
    }

    fn compare(&self, stored_hash: &str, password: &str) -> Result<(), PasswordError> {
        match bcrypt::verify(password, stored_hash) {
            Ok(true) => Ok(()),
            Ok(false) => Err(PasswordError::Mismatch),
            Err(source) => Err(PasswordError::CompareFailure {
                source: Box::new(source),
            }),
        }
    }
}
