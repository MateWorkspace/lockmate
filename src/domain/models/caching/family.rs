/// Invalidation boundaries. Space-scoped IDs must be positive.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash)]
pub enum CacheFamily {
    Users,
    Spaces,
    Permissions { space_id: i64 },
    Roles { space_id: i64 },
    SpaceMembers { space_id: i64 },
    RolePermissions { space_id: i64 },
    MemberRoles { space_id: i64 },
    ApiKeys { space_id: i64 },
}

/// A fresh, non-reusable generation; missing metadata must never imply zero.
#[derive(Clone, PartialEq, Eq)]
pub struct CacheRevision {
    pub family: CacheFamily,
    pub token: String,
}
