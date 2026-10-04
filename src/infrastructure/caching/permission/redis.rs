use std::time::Duration;

use crate::domain::contracts::repository::PermissionFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, Permission as PermissionEntity,
};

use super::super::{RedisBackend, shared::query::PayloadKind};
use crate::domain::contracts::caching::{CachingFuture, Permission};
use crate::domain::models::CacheQuery;

pub struct RedisPermission {
    rdb: RedisBackend,
}

impl RedisPermission {
    pub fn new(rdb: RedisBackend) -> Self {
        Self { rdb }
    }
}

impl Permission for RedisPermission {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<PermissionEntity>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::PermissionById { space_id, id },
                    (
                        PayloadKind::PermissionRecord,
                        "caching/permission/redis/read_by_id",
                    ),
                )
                .await
        })
    }

    fn read_by_slug<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        slug: &'a str,
    ) -> CachingFuture<'a, CacheRead<PermissionEntity>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::PermissionBySlug {
                        space_id,
                        slug: slug.to_owned(),
                    },
                    (
                        PayloadKind::PermissionRecord,
                        "caching/permission/redis/read_by_slug",
                    ),
                )
                .await
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        filter: PermissionFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<PermissionEntity>>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::PermissionByFilter { space_id, filter },
                    (
                        PayloadKind::PermissionPage,
                        "caching/permission/redis/read_by_filter",
                    ),
                )
                .await
        })
    }

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a PermissionEntity,
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
                        PayloadKind::PermissionRecord,
                        "caching/permission/redis/store_record",
                    ),
                )
                .await
        })
    }

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<PermissionEntity>,
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
                        PayloadKind::PermissionPage,
                        "caching/permission/redis/store_page",
                    ),
                )
                .await
        })
    }
}
