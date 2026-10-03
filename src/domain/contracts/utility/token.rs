use crate::domain::models::{AccessClaims, AppContext, RefreshClaims, TokenError};

pub trait Token: Send + Sync {
    fn generate_access(
        &self,
        context: &AppContext,
        claims: &AccessClaims,
    ) -> Result<String, TokenError>;

    fn validate_access(
        &self,
        context: &AppContext,
        token: &str,
    ) -> Result<AccessClaims, TokenError>;

    fn generate_refresh(
        &self,
        context: &AppContext,
        claims: &RefreshClaims,
    ) -> Result<String, TokenError>;

    fn validate_refresh(
        &self,
        context: &AppContext,
        token: &str,
    ) -> Result<RefreshClaims, TokenError>;
}
