use std::time::Duration;

use crate::domain::contracts::repository::PermissionFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, Permission as PermissionEntity,
};

use super::CachingFuture;

/// Management cache only; bypass for authentication and transactional reads.
pub trait Permission: Send + Sync {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<PermissionEntity>>;

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        slug: &'a str,
    ) -> CachingFuture<'a, CacheRead<PermissionEntity>>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: PermissionFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<PermissionEntity>>>;

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a PermissionEntity,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<PermissionEntity>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;
}
