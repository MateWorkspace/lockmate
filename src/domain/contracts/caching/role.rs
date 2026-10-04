use std::time::Duration;

use crate::domain::contracts::repository::RoleFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, Role as RoleEntity,
};

use super::CachingFuture;

/// Management cache only; bypass for authentication and transactional reads.
pub trait Role: Send + Sync {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<RoleEntity>>;

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        slug: &'a str,
    ) -> CachingFuture<'a, CacheRead<RoleEntity>>;

    fn read_default<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
    ) -> CachingFuture<'a, CacheRead<RoleEntity>>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: RoleFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<RoleEntity>>>;

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a RoleEntity,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<RoleEntity>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;
}
