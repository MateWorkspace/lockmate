use crate::domain::models::{AppContext, ValidatorError};

pub trait Validator: Send + Sync {
    fn permission_name(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;
    fn permission_desc(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;

    fn role_name(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;
    fn role_desc(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;

    fn user_name(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;
    fn user_bio(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;
    fn user_username(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;
    fn user_email(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;
    fn user_phone(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;
    fn user_password(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;

    fn api_key_name(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;
    fn api_key_desc(&self, context: &AppContext, value: &str) -> Result<(), ValidatorError>;
}
