use super::super::RedisBackend;
use crate::domain::{
    contracts::caching::{CachingFuture, Invalidation},
    models::{AppContext, CacheFamily},
};

pub struct RedisInvalidation {
    rdb: RedisBackend,
}

impl RedisInvalidation {
    pub fn new(rdb: RedisBackend) -> Self {
        Self { rdb }
    }
}

impl Invalidation for RedisInvalidation {
    fn invalidate<'a>(
        &'a self,
        context: &'a AppContext,
        families: &'a [CacheFamily],
    ) -> CachingFuture<'a, ()> {
        Box::pin(async move { self.rdb.invalidate(context, families).await })
    }
}
