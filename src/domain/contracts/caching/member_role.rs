use std::time::Duration;

use crate::domain::models::{
    AppContext, CacheRead, CacheStamp, CacheStoreOutcome, CachedMemberRoleDetails,
    CachedMemberRoleWithMember, CachedMemberRoleWithRole,
};

use super::CachingFuture;

/// Management cache only; bypass for authentication and transactional reads.
pub trait MemberRole: Send + Sync {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedMemberRoleDetails>>;

    fn read_by_member_id_and_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
        role_id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedMemberRoleDetails>>;

    fn read_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
    ) -> CachingFuture<'a, CacheRead<Vec<CachedMemberRoleWithRole>>>;

    fn read_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        role_id: i64,
    ) -> CachingFuture<'a, CacheRead<Vec<CachedMemberRoleWithMember>>>;

    fn store_details<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachedMemberRoleDetails,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;

    fn store_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a [CachedMemberRoleWithRole],
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;

    fn store_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a [CachedMemberRoleWithMember],
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome>;
}
