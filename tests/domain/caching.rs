use std::{error::Error, sync::Arc, time::Duration};

use lockmate::domain::{
    contracts::{
        caching::{self, CachingFuture},
        repository::*,
    },
    models::{self, *},
};
use models::{
    Permission as PermissionEntity, Role as RoleEntity, Space as SpaceEntity,
    SpaceMember as SpaceMemberEntity,
};

struct MockCache;

macro_rules! mock_cache {
    ($contract:ident { $($method:ident<$lt:lifetime>($($arg:ident: $ty:ty),*) -> $out:ty;)* }) => {
        impl caching::$contract for MockCache {
            $(fn $method<$lt>(&$lt self, $($arg: $ty),*) -> CachingFuture<$lt, $out> {
                let _ = ($($arg,)*);
                Box::pin(async { Err(CachingError::BadState) })
            })*
        }
    };
}

mock_cache!(Space {
    read_by_id<'a>(context: &'a AppContext, id: i64) -> CacheRead<SpaceEntity>;
    read_by_slug<'a>(context: &'a AppContext, slug: &'a str) -> CacheRead<SpaceEntity>;
    read_by_filter<'a>(context: &'a AppContext, filter: SpaceFilter) -> CacheRead<CachePage<SpaceEntity>>;
    store_record<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a SpaceEntity, ttl: Duration) -> CacheStoreOutcome;
    store_page<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a CachePage<SpaceEntity>, ttl: Duration) -> CacheStoreOutcome;
});

mock_cache!(User {
    read_by_id<'a>(context: &'a AppContext, id: i64) -> CacheRead<CachedUser>;
    read_by_username<'a>(context: &'a AppContext, username: &'a str) -> CacheRead<CachedUser>;
    read_by_email<'a>(context: &'a AppContext, email: &'a str) -> CacheRead<CachedUser>;
    read_by_phone<'a>(context: &'a AppContext, phone: &'a str) -> CacheRead<CachedUser>;
    read_by_filter<'a>(context: &'a AppContext, filter: UserFilter) -> CacheRead<CachePage<CachedUser>>;
    store_record<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a CachedUser, ttl: Duration) -> CacheStoreOutcome;
    store_page<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a CachePage<CachedUser>, ttl: Duration) -> CacheStoreOutcome;
});

mock_cache!(Permission {
    read_by_id<'a>(context: &'a AppContext, space_id: i64, id: i64) -> CacheRead<PermissionEntity>;
    read_by_slug<'a>(context: &'a AppContext, space_id: i64, slug: &'a str) -> CacheRead<PermissionEntity>;
    read_by_filter<'a>(context: &'a AppContext, space_id: i64, filter: PermissionFilter) -> CacheRead<CachePage<PermissionEntity>>;
    store_record<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a PermissionEntity, ttl: Duration) -> CacheStoreOutcome;
    store_page<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a CachePage<PermissionEntity>, ttl: Duration) -> CacheStoreOutcome;
});

mock_cache!(Role {
    read_by_id<'a>(context: &'a AppContext, space_id: i64, id: i64) -> CacheRead<RoleEntity>;
    read_by_slug<'a>(context: &'a AppContext, space_id: i64, slug: &'a str) -> CacheRead<RoleEntity>;
    read_default<'a>(context: &'a AppContext, space_id: i64) -> CacheRead<RoleEntity>;
    read_by_filter<'a>(context: &'a AppContext, space_id: i64, filter: RoleFilter) -> CacheRead<CachePage<RoleEntity>>;
    store_record<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a RoleEntity, ttl: Duration) -> CacheStoreOutcome;
    store_page<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a CachePage<RoleEntity>, ttl: Duration) -> CacheStoreOutcome;
});

mock_cache!(SpaceMember {
    read_by_id<'a>(context: &'a AppContext, space_id: i64, id: i64) -> CacheRead<SpaceMemberEntity>;
    read_by_user_id<'a>(context: &'a AppContext, space_id: i64, user_id: i64) -> CacheRead<SpaceMemberEntity>;
    read_by_filter<'a>(context: &'a AppContext, space_id: i64, filter: SpaceMemberFilter) -> CacheRead<CachePage<CachedSpaceMemberWithUser>>;
    store_record<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a SpaceMemberEntity, ttl: Duration) -> CacheStoreOutcome;
    store_page<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a CachePage<CachedSpaceMemberWithUser>, ttl: Duration) -> CacheStoreOutcome;
});

mock_cache!(RolePermission {
    read_by_id<'a>(context: &'a AppContext, space_id: i64, id: i64) -> CacheRead<CachedRolePermissionDetails>;
    read_by_role_id_and_permission_id<'a>(context: &'a AppContext, space_id: i64, role_id: i64, permission_id: i64) -> CacheRead<CachedRolePermissionDetails>;
    read_by_role_id<'a>(context: &'a AppContext, space_id: i64, role_id: i64) -> CacheRead<Vec<CachedRolePermissionWithPermission>>;
    read_by_permission_id<'a>(context: &'a AppContext, space_id: i64, permission_id: i64) -> CacheRead<Vec<CachedRolePermissionWithRole>>;
    store_details<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a CachedRolePermissionDetails, ttl: Duration) -> CacheStoreOutcome;
    store_by_role_id<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a [CachedRolePermissionWithPermission], ttl: Duration) -> CacheStoreOutcome;
    store_by_permission_id<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a [CachedRolePermissionWithRole], ttl: Duration) -> CacheStoreOutcome;
});

mock_cache!(MemberRole {
    read_by_id<'a>(context: &'a AppContext, space_id: i64, id: i64) -> CacheRead<CachedMemberRoleDetails>;
    read_by_member_id_and_role_id<'a>(context: &'a AppContext, space_id: i64, member_id: i64, role_id: i64) -> CacheRead<CachedMemberRoleDetails>;
    read_by_member_id<'a>(context: &'a AppContext, space_id: i64, member_id: i64) -> CacheRead<Vec<CachedMemberRoleWithRole>>;
    read_by_role_id<'a>(context: &'a AppContext, space_id: i64, role_id: i64) -> CacheRead<Vec<CachedMemberRoleWithMember>>;
    store_details<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a CachedMemberRoleDetails, ttl: Duration) -> CacheStoreOutcome;
    store_by_member_id<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a [CachedMemberRoleWithRole], ttl: Duration) -> CacheStoreOutcome;
    store_by_role_id<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a [CachedMemberRoleWithMember], ttl: Duration) -> CacheStoreOutcome;
});

mock_cache!(ApiKey {
    read_by_id<'a>(context: &'a AppContext, space_id: i64, id: i64) -> CacheRead<CachedApiKey>;
    read_by_filter<'a>(context: &'a AppContext, space_id: i64, filter: ApiKeyFilter) -> CacheRead<CachePage<CachedApiKey>>;
    store_record<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a CachedApiKey, ttl: Duration) -> CacheStoreOutcome;
    store_page<'a>(context: &'a AppContext, stamp: &'a CacheStamp, value: &'a CachePage<CachedApiKey>, ttl: Duration) -> CacheStoreOutcome;
});

impl caching::Invalidation for MockCache {
    fn invalidate<'a>(&'a self, _: &'a AppContext, _: &'a [CacheFamily]) -> CachingFuture<'a, ()> {
        Box::pin(async { Ok(()) })
    }
}

impl caching::KeyBuilder for MockCache {
    fn entry(&self, _: &AppContext, _: &CacheEntryKey) -> Result<String, CachingError> {
        Err(CachingError::BadArgs)
    }
    fn revision(&self, _: &AppContext, _: CacheFamily) -> Result<String, CachingError> {
        Err(CachingError::BadArgs)
    }
}

#[tokio::test]
async fn contracts_support_shared_trait_objects_and_send_futures() {
    let context = AppContext::default();
    let _: Arc<dyn caching::Space> = Arc::new(MockCache);
    let _: Arc<dyn caching::User> = Arc::new(MockCache);
    let _: Arc<dyn caching::Permission> = Arc::new(MockCache);
    let _: Arc<dyn caching::Role> = Arc::new(MockCache);
    let _: Arc<dyn caching::SpaceMember> = Arc::new(MockCache);
    let _: Arc<dyn caching::RolePermission> = Arc::new(MockCache);
    let _: Arc<dyn caching::MemberRole> = Arc::new(MockCache);
    let cache: Arc<dyn caching::ApiKey> = Arc::new(MockCache);
    let invalidation: Arc<dyn caching::Invalidation> = Arc::new(MockCache);
    let keys: Arc<dyn caching::KeyBuilder> = Arc::new(MockCache);
    fn assert_send<T: Send>(_: &T) {}
    let future = cache.read_by_id(&context, 1, 1);
    assert_send(&future);
    assert!(matches!(future.await, Err(CachingError::BadState)));
    assert!(invalidation.invalidate(&context, &[]).await.is_ok());
    assert!(matches!(
        keys.revision(&context, CacheFamily::Users),
        Err(CachingError::BadArgs)
    ));
}

#[test]
fn caching_errors_have_safe_messages_codes_and_preserve_sources() {
    let cases = [
        (
            CachingError::BadArgs,
            "BAD_ARGS",
            "invalid arguments",
            false,
        ),
        (CachingError::BadState, "BAD_STATE", "invalid state", false),
        (
            CachingError::Timeout {
                source: Box::new(std::io::Error::other("private backend detail")),
            },
            "TIMEOUT",
            "caching operation timed out",
            true,
        ),
        (
            CachingError::Failure {
                source: Box::new(std::io::Error::other("private backend detail")),
            },
            "FAILURE",
            "caching operation failed",
            true,
        ),
    ];
    for (error, code, message, has_source) in cases {
        assert_eq!(error.code(), code);
        assert_eq!(error.to_string(), message);
        assert_eq!(error.source().is_some(), has_source);
        if let Some(source) = error.source() {
            assert_eq!(source.to_string(), "private backend detail");
        }
    }
}

fn audit() -> AuditCreateUpdateDelete {
    AuditCreateUpdateDelete {
        create: AuditCreate {
            at: time::OffsetDateTime::UNIX_EPOCH,
            by: None,
        },
        update: AuditUpdate::default(),
        delete: AuditDelete::default(),
    }
}

#[test]
fn management_projections_discard_credentials_including_joined_pages() {
    let user = models::User {
        id: 7,
        name: "Example User".into(),
        bio: String::new(),
        username: "example".into(),
        email: Some("example@example.com".into()),
        phone: None,
        password_hash: "private-password-hash".into(),
        is_email_verified: false,
        is_phone_verified: false,
        avatar_path: None,
        preferences: serde_json::json!({}),
        audit: audit(),
    };
    let safe = CachedUser::from(user);
    let member = SpaceMemberEntity {
        id: 8,
        space_id: 1,
        user_id: 7,
        is_active: true,
        preferences: serde_json::json!({}),
        audit: audit(),
    };
    let page = CachePage::from((vec![CachedSpaceMemberWithUser { member, user: safe }], 1));
    let value = serde_json::to_value(&page).unwrap();
    assert_eq!(value["total"], 1);
    assert_eq!(value["items"][0]["user"]["id"], 7);
    assert!(value["items"][0]["user"].get("password_hash").is_none());
    assert!(!value.to_string().contains("private-password-hash"));
    let decoded: CachePage<CachedSpaceMemberWithUser> = serde_json::from_value(value).unwrap();
    assert_eq!(decoded, page);

    let key = models::ApiKey {
        id: 9,
        space_id: 1,
        member_id: 8,
        name: "Example key".into(),
        description: String::new(),
        hash: "private-api-key-hash".into(),
        redacted: "abcd".into(),
        preferences: serde_json::json!({}),
        audit: audit(),
    };
    let safe = CachedApiKey::from(key);
    let value = serde_json::to_value(&safe).unwrap();
    assert!(value.get("hash").is_none());
    assert!(!value.to_string().contains("private-api-key-hash"));
    assert_eq!(serde_json::from_value::<CachedApiKey>(value).unwrap(), safe);
}
