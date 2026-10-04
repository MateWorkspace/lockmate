use std::time::Duration;

use crate::domain::contracts::repository::SpaceFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, Space as SpaceEntity,
};

use super::CachingFuture;

/// Management cache only; bypass for authentication and transactional reads.
pub trait Space: Send + Sync {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<SpaceEntity>>;

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        slug: &'a str,
    ) -> CachingFuture<'a, CacheRead<SpaceEntity>>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: SpaceFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<SpaceEntity>>>;

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a SpaceEntity,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<SpaceEntity>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;
}
