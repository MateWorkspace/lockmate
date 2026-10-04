use crate::domain::models::{AccessClaims, AppContext, RefreshClaims, TokenError};

pub trait Token: Send + Sync {
    fn generate_access(
        &self,
        context: &AppContext,
        claims: &AccessClaims,
    ) -> Result<String, TokenError>;

    /// Checks token contents against the caller's expected space, without repository reads.
    fn validate_access(
        &self,
        context: &AppContext,
        space_id: i64,
        token: &str,
    ) -> Result<AccessClaims, TokenError>;

    fn generate_refresh(
        &self,
        context: &AppContext,
        claims: &RefreshClaims,
    ) -> Result<String, TokenError>;

    /// Checks refresh purpose and scope; application code rechecks live membership.
    fn validate_refresh(
        &self,
        context: &AppContext,
        space_id: i64,
        token: &str,
    ) -> Result<RefreshClaims, TokenError>;
}
