use std::time::Duration;

use crate::domain::contracts::repository::UserFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, CachedUser,
};

use super::CachingFuture;

/// Management cache only; bypass for authentication and transactional reads.
pub trait User: Send + Sync {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedUser>>;

    fn read_by_username<'a>(
        &'a self,
        context: &'a AppContext,
        username: &'a str,
    ) -> CachingFuture<'a, CacheRead<CachedUser>>;

    fn read_by_email<'a>(
        &'a self,
        context: &'a AppContext,
        email: &'a str,
    ) -> CachingFuture<'a, CacheRead<CachedUser>>;

    fn read_by_phone<'a>(
        &'a self,
        context: &'a AppContext,
        phone: &'a str,
    ) -> CachingFuture<'a, CacheRead<CachedUser>>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: UserFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<CachedUser>>>;

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachedUser,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<CachedUser>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;
}
