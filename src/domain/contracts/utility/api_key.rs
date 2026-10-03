use crate::domain::models::{ApiKeyError, AppContext, GeneratedApiKey};

pub trait ApiKey: Send + Sync {
    fn generate(&self, context: &AppContext) -> Result<GeneratedApiKey, ApiKeyError>;

    fn hash(&self, context: &AppContext, raw: &str) -> String;

    fn validate(&self, context: &AppContext, raw: &str) -> Result<(), ApiKeyError>;
}
