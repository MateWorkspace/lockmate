use super::CacheEntryKey;

/// A miss still carries a stamp so a later database result can be stored safely.
#[derive(Clone, PartialEq)]
pub struct CacheRead<T> {
    pub value: Option<T>,
    pub stamp: CacheStamp,
}

/// Identifies the exact lookup and revisions observed by the cache read.
/// Treat as adapter-issued data; callers pass it back without modification.
#[derive(Clone, PartialEq)]
pub struct CacheStamp {
    pub key: CacheEntryKey,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CacheStoreOutcome {
    Stored,
    Superseded,
}
