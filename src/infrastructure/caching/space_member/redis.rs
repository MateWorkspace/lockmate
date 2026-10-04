use std::time::Duration;

use crate::domain::contracts::repository::SpaceMemberFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, CachedSpaceMemberWithUser,
    SpaceMember as SpaceMemberEntity,
};

use super::super::{RedisBackend, shared::query::PayloadKind};
use crate::domain::contracts::caching::{CachingFuture, SpaceMember};
use crate::domain::models::CacheQuery;

pub struct RedisSpaceMember {
    rdb: RedisBackend,
}

impl RedisSpaceMember {
    pub fn new(rdb: RedisBackend) -> Self {
        Self { rdb }
    }
}

impl SpaceMember for RedisSpaceMember {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<SpaceMemberEntity>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::SpaceMemberById { space_id, id },
                    (
                        PayloadKind::SpaceMemberRecord,
                        "caching/space_member/redis/read_by_id",
                    ),
                )
                .await
        })
    }

    fn read_by_user_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        user_id: i64,
    ) -> CachingFuture<'a, CacheRead<SpaceMemberEntity>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::SpaceMemberByUserId { space_id, user_id },
                    (
                        PayloadKind::SpaceMemberRecord,
                        "caching/space_member/redis/read_by_user_id",
                    ),
                )
                .await
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: SpaceMemberFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<CachedSpaceMemberWithUser>>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::SpaceMemberByFilter { space_id, filter },
                    (
                        PayloadKind::SpaceMemberPage,
                        "caching/space_member/redis/read_by_filter",
                    ),
                )
                .await
        })
    }

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a SpaceMemberEntity,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome> {
        Box::pin(async move {
            self.rdb
                .store(
                    context,
                    stamp,
                    value,
                    ttl,
                    (
                        PayloadKind::SpaceMemberRecord,
                        "caching/space_member/redis/store_record",
                    ),
                )
                .await
        })
    }

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<CachedSpaceMemberWithUser>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome> {
        Box::pin(async move {
            self.rdb
                .store(
                    context,
                    stamp,
                    value,
                    ttl,
                    (
                        PayloadKind::SpaceMemberPage,
                        "caching/space_member/redis/store_page",
                    ),
                )
                .await
        })
    }
}
