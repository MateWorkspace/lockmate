use std::time::Duration;

use crate::domain::contracts::repository::ApiKeyFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, CachedApiKey,
};

use super::CachingFuture;

/// Management cache only; bypass for authentication and transactional reads.
pub trait ApiKey: Send + Sync {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedApiKey>>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: ApiKeyFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<CachedApiKey>>>;

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachedApiKey,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<CachedApiKey>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;
}
