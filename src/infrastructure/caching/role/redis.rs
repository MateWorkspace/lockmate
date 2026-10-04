use std::time::Duration;

use crate::domain::contracts::repository::RoleFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, Role as RoleEntity,
};

use super::super::{RedisBackend, shared::query::PayloadKind};
use crate::domain::contracts::caching::{CachingFuture, Role};
use crate::domain::models::CacheQuery;

pub struct RedisRole {
    rdb: RedisBackend,
}

impl RedisRole {
    pub fn new(rdb: RedisBackend) -> Self {
        Self { rdb }
    }
}

impl Role for RedisRole {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<RoleEntity>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::RoleById { space_id, id },
                    (PayloadKind::RoleRecord, "caching/role/redis/read_by_id"),
                )
                .await
        })
    }

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        slug: &'a str,
    ) -> CachingFuture<'a, CacheRead<RoleEntity>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::RoleBySlug {
                        space_id,
                        slug: slug.to_owned(),
                    },
                    (PayloadKind::RoleRecord, "caching/role/redis/read_by_slug"),
                )
                .await
        })
    }

    fn read_default<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
    ) -> CachingFuture<'a, CacheRead<RoleEntity>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::RoleDefault { space_id },
                    (PayloadKind::RoleRecord, "caching/role/redis/read_default"),
                )
                .await
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: RoleFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<RoleEntity>>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::RoleByFilter { space_id, filter },
                    (PayloadKind::RolePage, "caching/role/redis/read_by_filter"),
                )
                .await
        })
    }

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a RoleEntity,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome> {
        Box::pin(async move {
            self.rdb
                .store(
                    context,
                    stamp,
                    value,
                    ttl,
                    (PayloadKind::RoleRecord, "caching/role/redis/store_record"),
                )
                .await
        })
    }

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<RoleEntity>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome> {
        Box::pin(async move {
            self.rdb
                .store(
                    context,
                    stamp,
                    value,
                    ttl,
                    (PayloadKind::RolePage, "caching/role/redis/store_page"),
                )
                .await
        })
    }
}
