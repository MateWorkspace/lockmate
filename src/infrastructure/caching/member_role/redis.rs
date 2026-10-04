use std::time::Duration;

use crate::domain::models::{
    AppContext, CacheRead, CacheStamp, CacheStoreOutcome, CachedMemberRoleDetails,
    CachedMemberRoleWithMember, CachedMemberRoleWithRole,
};

use super::super::{RedisBackend, shared::query::PayloadKind};
use crate::domain::contracts::caching::{CachingFuture, MemberRole};
use crate::domain::models::CacheQuery;

pub struct RedisMemberRole {
    rdb: RedisBackend,
}

impl RedisMemberRole {
    pub fn new(rdb: RedisBackend) -> Self {
        Self { rdb }
    }
}

impl MemberRole for RedisMemberRole {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedMemberRoleDetails>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::MemberRoleById { space_id, id },
                    (
                        PayloadKind::MemberRoleDetails,
                        "caching/member_role/redis/read_by_id",
                    ),
                )
                .await
        })
    }

    fn read_by_member_id_and_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
        role_id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedMemberRoleDetails>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::MemberRoleByPair {
                        space_id,
                        member_id,
                        role_id,
                    },
                    (
                        PayloadKind::MemberRoleDetails,
                        "caching/member_role/redis/read_by_member_id_and_role_id",
                    ),
                )
                .await
        })
    }

    fn read_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
    ) -> CachingFuture<'a, CacheRead<Vec<CachedMemberRoleWithRole>>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::MemberRoleByMemberId {
                        space_id,
                        member_id,
                    },
                    (
                        PayloadKind::MemberRoleWithRoleList,
                        "caching/member_role/redis/read_by_member_id",
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
    ) -> CachingFuture<'a, CacheRead<Vec<CachedMemberRoleWithMember>>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::MemberRoleByRoleId { space_id, role_id },
                    (
                        PayloadKind::MemberRoleWithMemberList,
                        "caching/member_role/redis/read_by_role_id",
                    ),
                )
                .await
        })
    }

    fn store_details<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachedMemberRoleDetails,
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
                        PayloadKind::MemberRoleDetails,
                        "caching/member_role/redis/store_details",
                    ),
                )
                .await
        })
    }

    fn store_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a [CachedMemberRoleWithRole],
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
                        PayloadKind::MemberRoleWithRoleList,
                        "caching/member_role/redis/store_by_member_id",
                    ),
                )
                .await
        })
    }

    fn store_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a [CachedMemberRoleWithMember],
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
                        PayloadKind::MemberRoleWithMemberList,
                        "caching/member_role/redis/store_by_role_id",
                    ),
                )
                .await
        })
    }
}
