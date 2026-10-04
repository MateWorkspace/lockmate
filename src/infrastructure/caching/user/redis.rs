use std::time::Duration;

use crate::domain::contracts::repository::UserFilter;
use crate::domain::models::{
    AppContext, CachePage, CacheRead, CacheStamp, CacheStoreOutcome, CachedUser,
};

use super::super::{RedisBackend, shared::query::PayloadKind};
use crate::domain::contracts::caching::{CachingFuture, User};
use crate::domain::models::CacheQuery;

pub struct RedisUser {
    rdb: RedisBackend,
}

impl RedisUser {
    pub fn new(rdb: RedisBackend) -> Self {
        Self { rdb }
    }
}

impl User for RedisUser {
    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        id: i64,
    ) -> CachingFuture<'a, CacheRead<CachedUser>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::UserById { id },
                    (PayloadKind::UserRecord, "caching/user/redis/read_by_id"),
                )
                .await
        })
    }

    fn read_by_username<'a>(
        &'a self,
        context: &'a AppContext,
        username: &'a str,
    ) -> CachingFuture<'a, CacheRead<CachedUser>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::UserByUsername {
                        username: username.to_owned(),
                    },
                    (
                        PayloadKind::UserRecord,
                        "caching/user/redis/read_by_username",
                    ),
                )
                .await
        })
    }

    fn read_by_email<'a>(
        &'a self,
        context: &'a AppContext,
        email: &'a str,
    ) -> CachingFuture<'a, CacheRead<CachedUser>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::UserByEmail {
                        email: email.to_owned(),
                    },
                    (PayloadKind::UserRecord, "caching/user/redis/read_by_email"),
                )
                .await
        })
    }

    fn read_by_phone<'a>(
        &'a self,
        context: &'a AppContext,
        phone: &'a str,
    ) -> CachingFuture<'a, CacheRead<CachedUser>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::UserByPhone {
                        phone: phone.to_owned(),
                    },
                    (PayloadKind::UserRecord, "caching/user/redis/read_by_phone"),
                )
                .await
        })
    }

    fn read_by_filter<'a>(
        &'a self,
        context: &'a AppContext,
        filter: UserFilter,
    ) -> CachingFuture<'a, CacheRead<CachePage<CachedUser>>> {
        Box::pin(async move {
            self.rdb
                .read(
                    context,
                    CacheQuery::UserByFilter { filter },
                    (PayloadKind::UserPage, "caching/user/redis/read_by_filter"),
                )
                .await
        })
    }

    fn store_record<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachedUser,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome> {
        Box::pin(async move {
            self.rdb
                .store(
                    context,
                    stamp,
                    value,
                    ttl,
                    (PayloadKind::UserRecord, "caching/user/redis/store_record"),
                )
                .await
        })
    }

    fn store_page<'a>(
        &'a self,
        context: &'a AppContext,
        stamp: &'a CacheStamp,
        value: &'a CachePage<CachedUser>,
        ttl: Duration,
    ) -> CachingFuture<'a, CacheStoreOutcome> {
        Box::pin(async move {
            self.rdb
                .store(
                    context,
                    stamp,
                    value,
                    ttl,
                    (PayloadKind::UserPage, "caching/user/redis/store_page"),
                )
                .await
        })
    }
}
