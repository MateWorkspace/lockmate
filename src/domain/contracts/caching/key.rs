use crate::domain::models::{AppContext, CacheEntryKey, CacheFamily, CachingError};

/// Builds versioned, namespaced keys without I/O. Namespace is constructor-injected.
pub trait KeyBuilder: Send + Sync {
    fn entry(
        &self,
        context: &AppContext,
        specification: &CacheEntryKey,
    ) -> Result<String, CachingError>;

    fn revision(&self, context: &AppContext, family: CacheFamily) -> Result<String, CachingError>;
}
