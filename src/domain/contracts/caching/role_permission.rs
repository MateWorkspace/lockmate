use std::time::Duration;

use crate::domain::models::{
    AppContext, CacheRead, CacheStamp, CacheStoreOutcome, CachedRolePermissionDetails,
    CachedRolePermissionWithPermission, CachedRolePermissionWithRole,
};

use super::CachingFuture;

/// Management cache only; bypass for authentication and transactional reads.
pub trait RolePermission: Send + Sync {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedRolePermissionDetails>>;

    fn read_by_role_id_and_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        role_id: i64,
        permission_id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedRolePermissionDetails>>;

    fn read_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        role_id: i64,
    ) -> CachingFuture<'a, CacheRead<Vec<CachedRolePermissionWithPermission>>>;

    fn read_by_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        permission_id: i64,
    ) -> CachingFuture<'a, CacheRead<Vec<CachedRolePermissionWithRole>>>;

    fn store_details<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachedRolePermissionDetails,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;

    fn store_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a [CachedRolePermissionWithPermission],
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;

    fn store_by_permission_id<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a [CachedRolePermissionWithRole],
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;
}
