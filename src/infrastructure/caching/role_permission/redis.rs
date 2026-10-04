use std::time::Duration;

use crate::domain::models::{
    AppContext, CacheRead, CacheStamp, CacheStoreOutcome, CachedRolePermissionDetails,
    CachedRolePermissionWithPermission, CachedRolePermissionWithRole,
};

use super::super::{RedisBackend, shared::query::PayloadKind};
use crate::domain::contracts::caching::{CachingFuture, RolePermission};
use crate::domain::models::CacheQuery;

pub struct RedisRolePermission {
    rdb: RedisBackend,
}

impl RedisRolePermission {
    pub fn new(rdb: RedisBackend) -> Self {
        Self { rdb }
    }
}

impl RolePermission for RedisRolePermission {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedRolePermissionDetails>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::RolePermissionById { space_id, id },
                    (
                        PayloadKind::RolePermissionDetails,
                        "caching/role_permission/redis/read_by_id",
                    ),
                )
                .await
        })
    }

    fn read_by_role_id_and_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        role_id: i64,
        permission_id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedRolePermissionDetails>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::RolePermissionByPair {
                        space_id,
                        role_id,
                        permission_id,
                    },
                    (
                        PayloadKind::RolePermissionDetails,
                        "caching/role_permission/redis/read_by_role_id_and_permission_id",
                    ),
                )
                .await
        })
    }

    fn read_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        role_id: i64,
    ) -> CachingFuture<'a, CacheRead<Vec<CachedRolePermissionWithPermission>>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::RolePermissionByRoleId { space_id, role_id },
                    (
                        PayloadKind::RolePermissionWithPermissionList,
                        "caching/role_permission/redis/read_by_role_id",
                    ),
                )
                .await
        })
    }

    fn read_by_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        permission_id: i64,
    ) -> CachingFuture<'a, CacheRead<Vec<CachedRolePermissionWithRole>>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::RolePermissionByPermissionId {
                        space_id,
                        permission_id,
                    },
                    (
                        PayloadKind::RolePermissionWithRoleList,
                        "caching/role_permission/redis/read_by_permission_id",
                    ),
                )
                .await
        })
    }

    fn store_details<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachedRolePermissionDetails,
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
                        PayloadKind::RolePermissionDetails,
                        "caching/role_permission/redis/store_details",
                    ),
                )
                .await
        })
    }

    fn store_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a [CachedRolePermissionWithPermission],
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
                        PayloadKind::RolePermissionWithPermissionList,
                        "caching/role_permission/redis/store_by_role_id",
                    ),
                )
                .await
        })
    }

    fn store_by_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a [CachedRolePermissionWithRole],
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
                        PayloadKind::RolePermissionWithRoleList,
                        "caching/role_permission/redis/store_by_permission_id",
                    ),
                )
                .await
        })
    }
}
