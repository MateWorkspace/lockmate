use crate::domain::models::{AppContext, PasswordError};

pub trait Password: Send + Sync {
    fn hash(&self, context: &AppContext, password: &str) -> Result<String, PasswordError>;

    fn compare(
        &self,
        context: &AppContext,
        stored_hash: &str,
        password: &str,
    ) -> Result<(), PasswordError>;
}
