use std::time::Duration;

use crate::domain::contracts::repository::SpaceFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, Space as SpaceEntity,
};

use super::super::{RedisBackend, shared::query::PayloadKind};
use crate::domain::contracts::caching::{CachingFuture, Space};
use crate::domain::models::CacheQuery;

pub struct RedisSpace {
    rdb: RedisBackend,
}

impl RedisSpace {
    pub fn new(rdb: RedisBackend) -> Self {
        Self { rdb }
    }
}

impl Space for RedisSpace {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<SpaceEntity>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::SpaceById { id },
                    (PayloadKind::SpaceRecord, "caching/space/redis/read_by_id"),
                )
                .await
        })
    }

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        slug: &'a str,
    ) -> CachingFuture<'a, CacheRead<SpaceEntity>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::SpaceBySlug {
                        slug: slug.to_owned(),
                    },
                    (PayloadKind::SpaceRecord, "caching/space/redis/read_by_slug"),
                )
                .await
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: SpaceFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<SpaceEntity>>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::SpaceByFilter { filter },
                    (PayloadKind::SpacePage, "caching/space/redis/read_by_filter"),
                )
                .await
        })
    }

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a SpaceEntity,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome> {
        Box::pin(async move {
            self.rdb
                .store(
                    context,
                    stamp,
                    value,
                    ttl,
                    (PayloadKind::SpaceRecord, "caching/space/redis/store_record"),
                )
                .await
        })
    }

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<SpaceEntity>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome> {
        Box::pin(async move {
            self.rdb
                .store(
                    context,
                    stamp,
                    value,
                    ttl,
                    (PayloadKind::SpacePage, "caching/space/redis/store_page"),
                )
                .await
        })
    }
}
