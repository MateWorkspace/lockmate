use crate::domain::models::PasswordError;

pub trait Password: Send + Sync {
    fn hash(&self, password: &str) -> Result<String, PasswordError>;

    fn compare(&self, stored_hash: &str, password: &str) -> Result<(), PasswordError>;
}
