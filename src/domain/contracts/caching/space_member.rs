use std::time::Duration;

use crate::domain::contracts::repository::SpaceMemberFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, CachedSpaceMemberWithUser,
    SpaceMember as SpaceMemberEntity,
};

use super::CachingFuture;

/// Management cache only; bypass for authentication and transactional reads.
pub trait SpaceMember: Send + Sync {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<SpaceMemberEntity>>;

    fn read_by_user_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        user_id: i64,
    ) -> CachingFuture<'a, CacheRead<SpaceMemberEntity>>;

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: SpaceMemberFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<CachedSpaceMemberWithUser>>>;

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a SpaceMemberEntity,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<CachedSpaceMemberWithUser>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;
}
