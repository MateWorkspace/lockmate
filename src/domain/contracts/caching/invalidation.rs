use crate::domain::models::{AppContext, CacheFamily};

use super::CachingFuture;

pub trait Invalidation: Send + Sync {
    /// Atomically replace all requested family revisions after the outer commit.
    /// Duplicate families are harmless; an empty slice is a no-op.
    fn invalidate<'a>(
        &'a self,
        context: &'a AppContext,
        families: &'a [CacheFamily],
    ) -> CachingFuture<'a, ()>;
}
