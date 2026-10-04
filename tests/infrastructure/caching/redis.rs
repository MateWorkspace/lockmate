use super::support::{self, Fixture};
use lockmate::{
    domain::{
        contracts::{
            caching::{self as cache, Invalidation},
            repository as repo,
        },
        models::*,
    },
    infrastructure::caching::*,
};
use std::{sync::Arc, time::Duration};

macro_rules! roundtrip {
    ($f:ident, $cache:ident, $read:ident($($arg:expr),*), $store:ident, $value:expr) => {{
        let value = $value;
        let missed = $cache.$read(&$f.context, $($arg),*).await.unwrap();
        assert!(missed.value.is_none());
        $f.track(&missed.stamp);
        assert_eq!($cache.$store(&$f.context, &missed.stamp, &value, Duration::from_secs(2)).await.unwrap(), CacheStoreOutcome::Stored);
        let hit = $cache.$read(&$f.context, $($arg),*).await.unwrap();
        assert_eq!(hit.value.unwrap(), value);
        // Every declared dependency must make this entry unreachable.
        for revision in missed.stamp.key.revisions {
            RedisInvalidation::new($f.backend.clone()).invalidate(&$f.context, &[revision.family]).await.unwrap();
            let changed = $cache.$read(&$f.context, $($arg),*).await.unwrap();
            assert!(changed.value.is_none());
            $f.track(&changed.stamp);
        }
    }};
}

#[tokio::test]
#[ignore = "requires isolated Redis via LOCKMATE_TEST_REDIS_URL"]
async fn every_typed_read_store_and_dependency_round_trips() {
    let f = Fixture::new().await;
    let space: Arc<dyn cache::Space> = Arc::new(RedisSpace::new(f.backend.clone()));
    let user: Arc<dyn cache::User> = Arc::new(RedisUser::new(f.backend.clone()));
    let permission: Arc<dyn cache::Permission> = Arc::new(RedisPermission::new(f.backend.clone()));
    let role: Arc<dyn cache::Role> = Arc::new(RedisRole::new(f.backend.clone()));
    let space_member: Arc<dyn cache::SpaceMember> =
        Arc::new(RedisSpaceMember::new(f.backend.clone()));
    let api_key: Arc<dyn cache::ApiKey> = Arc::new(RedisApiKey::new(f.backend.clone()));
    let role_permission: Arc<dyn cache::RolePermission> =
        Arc::new(RedisRolePermission::new(f.backend.clone()));
    let member_role: Arc<dyn cache::MemberRole> = Arc::new(RedisMemberRole::new(f.backend.clone()));
    let member_with_user = CachedSpaceMemberWithUser {
        member: support::member(),
        user: support::user(),
    };
    let rp_details = CachedRolePermissionDetails {
        role_permission: support::role_permission(),
        role: support::role(),
        permission: support::permission(),
    };
    let rp_with_permission = CachedRolePermissionWithPermission {
        role_permission: support::role_permission(),
        permission: support::permission(),
    };
    let rp_with_role = CachedRolePermissionWithRole {
        role_permission: support::role_permission(),
        role: support::role(),
    };
    let mr_details = CachedMemberRoleDetails {
        member_role: support::member_role(),
        member: support::member(),
        role: support::role(),
    };
    let mr_with_member = CachedMemberRoleWithMember {
        member_role: support::member_role(),
        member: support::member(),
    };
    let mr_with_role = CachedMemberRoleWithRole {
        member_role: support::member_role(),
        role: support::role(),
    };
    roundtrip!(f, space, read_by_id(1), store_record, support::space());
    roundtrip!(
        f,
        space,
        read_by_slug("example"),
        store_record,
        support::space()
    );
    roundtrip!(
        f,
        space,
        read_by_filter(repo::SpaceFilter {
            page: 1,
            limit: 10,
            ..Default::default()
        }),
        store_page,
        CachePage {
            items: vec![support::space()],
            total: 1
        }
    );
    roundtrip!(f, user, read_by_id(2), store_record, support::user());
    roundtrip!(
        f,
        user,
        read_by_username("example"),
        store_record,
        support::user()
    );
    roundtrip!(
        f,
        user,
        read_by_email("example@example.com"),
        store_record,
        support::user()
    );
    roundtrip!(
        f,
        user,
        read_by_phone("1234567890"),
        store_record,
        support::user()
    );
    roundtrip!(
        f,
        user,
        read_by_filter(repo::UserFilter {
            page: 1,
            limit: 10,
            ..Default::default()
        }),
        store_page,
        CachePage {
            items: vec![support::user()],
            total: 1
        }
    );
    roundtrip!(
        f,
        permission,
        read_by_id(1, 3),
        store_record,
        support::permission()
    );
    roundtrip!(
        f,
        permission,
        read_by_slug(1, "example.read"),
        store_record,
        support::permission()
    );
    roundtrip!(
        f,
        permission,
        read_by_filter(
            1,
            repo::PermissionFilter {
                page: 1,
                limit: 10,
                ..Default::default()
            }
        ),
        store_page,
        CachePage {
            items: vec![support::permission()],
            total: 1
        }
    );
    roundtrip!(f, role, read_by_id(1, 4), store_record, support::role());
    roundtrip!(
        f,
        role,
        read_by_slug(1, "example"),
        store_record,
        support::role()
    );
    roundtrip!(f, role, read_default(1), store_record, support::role());
    roundtrip!(
        f,
        role,
        read_by_filter(
            1,
            repo::RoleFilter {
                page: 1,
                limit: 10,
                ..Default::default()
            }
        ),
        store_page,
        CachePage {
            items: vec![support::role()],
            total: 1
        }
    );
    roundtrip!(
        f,
        space_member,
        read_by_id(1, 5),
        store_record,
        support::member()
    );
    roundtrip!(
        f,
        space_member,
        read_by_user_id(1, 2),
        store_record,
        support::member()
    );
    roundtrip!(
        f,
        space_member,
        read_by_filter(
            1,
            repo::SpaceMemberFilter {
                page: 1,
                limit: 10,
                ..Default::default()
            }
        ),
        store_page,
        CachePage {
            items: vec![member_with_user.clone()],
            total: 1
        }
    );
    roundtrip!(
        f,
        role_permission,
        read_by_id(1, 7),
        store_details,
        rp_details.clone()
    );
    roundtrip!(
        f,
        role_permission,
        read_by_role_id_and_permission_id(1, 4, 3),
        store_details,
        rp_details.clone()
    );
    roundtrip!(
        f,
        role_permission,
        read_by_role_id(1, 4),
        store_by_role_id,
        vec![rp_with_permission.clone()]
    );
    roundtrip!(
        f,
        role_permission,
        read_by_permission_id(1, 3),
        store_by_permission_id,
        vec![rp_with_role.clone()]
    );
    roundtrip!(
        f,
        member_role,
        read_by_id(1, 8),
        store_details,
        mr_details.clone()
    );
    roundtrip!(
        f,
        member_role,
        read_by_member_id_and_role_id(1, 5, 4),
        store_details,
        mr_details.clone()
    );
    roundtrip!(
        f,
        member_role,
        read_by_member_id(1, 5),
        store_by_member_id,
        vec![mr_with_role.clone()]
    );
    roundtrip!(
        f,
        member_role,
        read_by_role_id(1, 4),
        store_by_role_id,
        vec![mr_with_member.clone()]
    );
    roundtrip!(
        f,
        api_key,
        read_by_id(1, 6),
        store_record,
        support::api_key()
    );
    roundtrip!(
        f,
        api_key,
        read_by_filter(
            1,
            repo::ApiKeyFilter {
                page: 1,
                limit: 10,
                ..Default::default()
            }
        ),
        store_page,
        CachePage {
            items: vec![support::api_key()],
            total: 1
        }
    );
    assert!(f.logger.0.lock().unwrap().is_empty());
    f.cleanup().await;
}

#[tokio::test]
#[ignore = "requires isolated Redis via LOCKMATE_TEST_REDIS_URL"]
async fn stale_fills_and_revision_loss_cannot_resurrect_entries() {
    let f = Fixture::new().await;
    let cache: Arc<dyn cache::Space> = Arc::new(RedisSpace::new(f.backend.clone()));
    let invalidation = RedisInvalidation::new(f.backend.clone());
    let before = cache.read_by_id(&f.context, 1).await.unwrap();
    let old_key = f.track(&before.stamp);
    cache
        .store_record(
            &f.context,
            &before.stamp,
            &support::space(),
            Duration::from_secs(10),
        )
        .await
        .unwrap();
    invalidation
        .invalidate(&f.context, &[CacheFamily::Spaces, CacheFamily::Spaces])
        .await
        .unwrap();
    assert_eq!(
        cache
            .store_record(
                &f.context,
                &before.stamp,
                &support::space(),
                Duration::from_secs(10)
            )
            .await
            .unwrap(),
        CacheStoreOutcome::Superseded
    );
    let after = cache.read_by_id(&f.context, 1).await.unwrap();
    f.track(&after.stamp);
    assert!(after.value.is_none());
    assert!(before.stamp != after.stamp);
    let old_bytes: Option<Vec<u8>> = redis::cmd("GET")
        .arg(old_key)
        .query_async(&mut f.raw.clone())
        .await
        .unwrap();
    assert!(old_bytes.is_some());
    cache
        .store_record(
            &f.context,
            &after.stamp,
            &support::space(),
            Duration::from_secs(10),
        )
        .await
        .unwrap();
    let revision_key = f.revision(CacheFamily::Spaces);
    redis::cmd("DEL")
        .arg(&revision_key)
        .query_async::<i64>(&mut f.raw.clone())
        .await
        .unwrap();
    assert_eq!(
        cache
            .store_record(
                &f.context,
                &after.stamp,
                &support::space(),
                Duration::from_secs(10)
            )
            .await
            .unwrap(),
        CacheStoreOutcome::Superseded
    );
    let recreated = cache.read_by_id(&f.context, 1).await.unwrap();
    f.track(&recreated.stamp);
    assert!(recreated.value.is_none());
    assert!(recreated.stamp != after.stamp);
    assert_eq!(
        redis::cmd("PTTL")
            .arg(revision_key)
            .query_async::<i64>(&mut f.raw.clone())
            .await
            .unwrap(),
        -1
    );
    invalidation.invalidate(&f.context, &[]).await.unwrap();
    assert!(f.logger.0.lock().unwrap().is_empty());
    f.cleanup().await;
}

#[tokio::test]
#[ignore = "requires isolated Redis via LOCKMATE_TEST_REDIS_URL"]
async fn pages_empty_lists_ttl_expiry_and_script_reload_work() {
    let f = Fixture::new().await;
    let spaces: Arc<dyn cache::Space> = Arc::new(RedisSpace::new(f.backend.clone()));
    let assignments: Arc<dyn cache::MemberRole> = Arc::new(RedisMemberRole::new(f.backend.clone()));
    let filter = repo::SpaceFilter {
        page: 1,
        limit: 10,
        ..Default::default()
    };
    let missed = spaces
        .read_by_filter(&f.context, filter.clone())
        .await
        .unwrap();
    let entry = f.track(&missed.stamp);
    let page = CachePage {
        items: vec![],
        total: 0,
    };
    spaces
        .store_page(&f.context, &missed.stamp, &page, Duration::from_millis(50))
        .await
        .unwrap();
    assert_eq!(
        spaces
            .read_by_filter(&f.context, filter.clone())
            .await
            .unwrap()
            .value
            .unwrap(),
        page
    );
    tokio::time::sleep(Duration::from_millis(100)).await;
    assert!(
        spaces
            .read_by_filter(&f.context, filter)
            .await
            .unwrap()
            .value
            .is_none()
    );
    assert_eq!(
        redis::cmd("PTTL")
            .arg(entry)
            .query_async::<i64>(&mut f.raw.clone())
            .await
            .unwrap(),
        -2
    );
    let empty = assignments
        .read_by_member_id(&f.context, 1, 5)
        .await
        .unwrap();
    f.track(&empty.stamp);
    assignments
        .store_by_member_id(&f.context, &empty.stamp, &[], Duration::from_secs(2))
        .await
        .unwrap();
    assert_eq!(
        assignments
            .read_by_member_id(&f.context, 1, 5)
            .await
            .unwrap()
            .value
            .unwrap()
            .len(),
        0
    );
    // SCRIPT FLUSH changes global script metadata only; no user keys are removed.
    redis::cmd("SCRIPT")
        .arg("FLUSH")
        .query_async::<()>(&mut f.raw.clone())
        .await
        .unwrap();
    assert!(
        assignments
            .read_by_member_id(&f.context, 1, 5)
            .await
            .unwrap()
            .value
            .is_some()
    );
    let record = spaces.read_by_id(&f.context, 1).await.unwrap();
    f.track(&record.stamp);
    spaces
        .store_record(
            &f.context,
            &record.stamp,
            &support::space(),
            Duration::from_millis(50),
        )
        .await
        .unwrap();
    tokio::time::sleep(Duration::from_millis(100)).await;
    assert!(
        spaces
            .read_by_id(&f.context, 1)
            .await
            .unwrap()
            .value
            .is_none()
    );
    // A positive sub-millisecond TTL rounds up rather than being rejected as zero.
    assert_eq!(
        spaces
            .store_record(
                &f.context,
                &record.stamp,
                &support::space(),
                Duration::from_nanos(1)
            )
            .await
            .unwrap(),
        CacheStoreOutcome::Stored
    );
    f.cleanup().await;
}

#[tokio::test]
#[ignore = "requires isolated Redis via LOCKMATE_TEST_REDIS_URL"]
async fn corrupt_payloads_invalid_inputs_and_transactions_are_safe_and_log_once() {
    let f = Fixture::new().await;
    let spaces: Arc<dyn cache::Space> = Arc::new(RedisSpace::new(f.backend.clone()));
    let users: Arc<dyn cache::User> = Arc::new(RedisUser::new(f.backend.clone()));
    let roles: Arc<dyn cache::Role> = Arc::new(RedisRole::new(f.backend.clone()));
    let record = spaces.read_by_id(&f.context, 1).await.unwrap();
    let key = f.track(&record.stamp);
    assert!(matches!(
        spaces
            .store_record(&f.context, &record.stamp, &support::space(), Duration::ZERO)
            .await,
        Err(CachingError::BadArgs)
    ));
    assert!(matches!(
        spaces
            .store_record(&f.context, &record.stamp, &support::space(), Duration::MAX)
            .await,
        Err(CachingError::BadArgs)
    ));
    let mut wrong = support::space();
    wrong.id = 2;
    assert!(matches!(
        spaces
            .store_record(&f.context, &record.stamp, &wrong, Duration::from_secs(1))
            .await,
        Err(CachingError::BadArgs)
    ));
    assert!(matches!(
        spaces
            .store_page(
                &f.context,
                &record.stamp,
                &CachePage {
                    items: vec![],
                    total: 0
                },
                Duration::from_secs(1)
            )
            .await,
        Err(CachingError::BadArgs)
    ));
    let transaction = AppContext {
        transaction: Some(AppTransaction::new(())),
        ..f.context.clone()
    };
    assert!(matches!(
        spaces.read_by_id(&transaction, 1).await,
        Err(CachingError::BadState)
    ));
    assert!(matches!(
        RedisInvalidation::new(f.backend.clone())
            .invalidate(&transaction, &[])
            .await,
        Err(CachingError::BadState)
    ));
    redis::cmd("SET")
        .arg(&key)
        .arg("private-broken-payload")
        .query_async::<()>(&mut f.raw.clone())
        .await
        .unwrap();
    assert!(matches!(
        spaces.read_by_id(&f.context, 1).await,
        Err(CachingError::Failure { .. })
    ));
    let invalid = serde_json::to_vec(&wrong).unwrap();
    redis::cmd("SET")
        .arg(&key)
        .arg(invalid)
        .query_async::<()>(&mut f.raw.clone())
        .await
        .unwrap();
    assert!(matches!(
        spaces.read_by_id(&f.context, 1).await,
        Err(CachingError::Failure { .. })
    ));
    let defaults = roles.read_default(&f.context, 1).await.unwrap();
    f.track(&defaults.stamp);
    let mut role = support::role();
    role.is_default = false;
    assert!(matches!(
        roles
            .store_record(&f.context, &defaults.stamp, &role, Duration::from_secs(1))
            .await,
        Err(CachingError::BadArgs)
    ));
    let user = users.read_by_id(&f.context, 2).await.unwrap();
    let user_key = f.track(&user.stamp);
    let mut secret = serde_json::to_value(support::user()).unwrap();
    secret["password_hash"] = "private-credential".into();
    redis::cmd("SET")
        .arg(user_key)
        .arg(serde_json::to_vec(&secret).unwrap())
        .query_async::<()>(&mut f.raw.clone())
        .await
        .unwrap();
    assert!(matches!(
        users.read_by_id(&f.context, 2).await,
        Err(CachingError::Failure { .. })
    ));
    {
        let entries = f.logger.0.lock().unwrap();
        assert_eq!(entries.len(), 10);
        for entry in entries.iter() {
            assert_eq!(entry.context.actor, f.context.actor);
            assert_eq!(entry.context.trace_id, f.context.trace_id);
            assert!(entry.tag.starts_with("caching/"));
            assert!(!entry.message.contains("private-"));
            assert!(!format!("{:?}", entry.meta).contains("private-"));
        }
        assert_eq!(entries[6].level, LoggerLevel::Warn);
    }
    f.cleanup().await;
}

#[tokio::test]
#[ignore = "requires isolated Redis via LOCKMATE_TEST_REDIS_URL"]
async fn revision_preflight_prevents_partial_writes_and_concurrent_initialization_agrees() {
    let f = Fixture::new().await;
    let members: Arc<dyn cache::SpaceMember> = Arc::new(RedisSpaceMember::new(f.backend.clone()));
    let concurrent =
        futures_util::future::join_all((0..32).map(|_| members.read_by_id(&f.context, 1, 5))).await;
    let first = concurrent[0].as_ref().unwrap();
    f.track(&first.stamp);
    for result in &concurrent {
        assert!(result.as_ref().unwrap().stamp == first.stamp);
    }
    let users_key = f.revision(CacheFamily::Users);
    let spaces_key = f.revision(CacheFamily::Spaces);
    let before: String = redis::cmd("GET")
        .arg(&users_key)
        .query_async(&mut f.raw.clone())
        .await
        .unwrap();
    redis::cmd("DEL")
        .arg(&spaces_key)
        .query_async::<i64>(&mut f.raw.clone())
        .await
        .unwrap();
    redis::cmd("RPUSH")
        .arg(&spaces_key)
        .arg("wrong-type")
        .query_async::<i64>(&mut f.raw.clone())
        .await
        .unwrap();
    let invalidation = RedisInvalidation::new(f.backend.clone());
    assert!(matches!(
        invalidation
            .invalidate(&f.context, &[CacheFamily::Users, CacheFamily::Spaces])
            .await,
        Err(CachingError::Failure { .. })
    ));
    let after: String = redis::cmd("GET")
        .arg(&users_key)
        .query_async(&mut f.raw.clone())
        .await
        .unwrap();
    assert_eq!(before, after);
    redis::cmd("DEL")
        .arg(&users_key)
        .query_async::<i64>(&mut f.raw.clone())
        .await
        .unwrap();
    assert!(matches!(
        members.read_by_id(&f.context, 1, 5).await,
        Err(CachingError::Failure { .. })
    ));
    let absent: Option<String> = redis::cmd("GET")
        .arg(&users_key)
        .query_async(&mut f.raw.clone())
        .await
        .unwrap();
    assert!(absent.is_none());
    redis::cmd("DEL")
        .arg(&spaces_key)
        .query_async::<i64>(&mut f.raw.clone())
        .await
        .unwrap();
    let recovered = members.read_by_id(&f.context, 1, 5).await.unwrap();
    f.track(&recovered.stamp);
    assert!(recovered.stamp != first.stamp);
    let before = recovered.stamp;
    invalidation
        .invalidate(&f.context, &[CacheFamily::Users, CacheFamily::Spaces])
        .await
        .unwrap();
    let after = members.read_by_id(&f.context, 1, 5).await.unwrap();
    f.track(&after.stamp);
    for family in [CacheFamily::Users, CacheFamily::Spaces] {
        assert!(
            before
                .key
                .revisions
                .iter()
                .find(|r| r.family == family)
                .unwrap()
                .token
                != after
                    .stamp
                    .key
                    .revisions
                    .iter()
                    .find(|r| r.family == family)
                    .unwrap()
                    .token
        );
    }
    f.cleanup().await;
}

#[tokio::test]
#[ignore = "requires isolated Redis via LOCKMATE_TEST_REDIS_URL"]
async fn authentication_failure_keeps_native_source_and_safe_single_log() {
    let mut f = Fixture::new().await;
    f.config.redis_password = "private-invalid-password".into();
    f.config.redis_connect_timeout = Duration::from_millis(300);
    let result =
        RedisBackend::connect(&f.context, &f.config, f.keys.clone(), f.logger.clone()).await;
    let Err(CachingError::Failure { source }) = result else {
        panic!("expected authentication failure")
    };
    assert!(source.downcast_ref::<redis::RedisError>().is_some());
    {
        let entries = f.logger.0.lock().unwrap();
        assert_eq!(entries.len(), 1);
        assert_eq!(entries[0].level, LoggerLevel::Warn);
        assert!(
            !format!("{} {:?}", entries[0].message, entries[0].meta)
                .contains("private-invalid-password")
        );
    }
    f.cleanup().await;
}

#[tokio::test]
#[ignore = "requires isolated Redis via LOCKMATE_TEST_REDIS_URL"]
async fn whole_operation_deadlines_and_connection_recovery_work() {
    let mut f = Fixture::new().await;
    let proxy = support::Proxy::new(f.config.redis_host.clone(), f.config.redis_port).await;
    f.config.redis_port = proxy.port;
    f.config.redis_operation_timeout = Duration::from_millis(50);
    let backend = RedisBackend::connect(&f.context, &f.config, f.keys.clone(), f.logger.clone())
        .await
        .unwrap();
    let cache: Arc<dyn cache::Space> = Arc::new(RedisSpace::new(backend));
    f.revision(CacheFamily::Spaces);
    proxy.pause();
    let before = tokio::time::Instant::now();
    let result = cache.read_by_id(&f.context, 1).await;
    assert!(matches!(result, Err(CachingError::Timeout { .. })));
    assert!(before.elapsed() < Duration::from_secs(1));
    {
        let entries = f.logger.0.lock().unwrap();
        assert_eq!(entries.len(), 1);
        assert_eq!(entries[0].level, LoggerLevel::Warn);
        assert_eq!(
            entries[0].meta.get("error_code"),
            Some(&LoggerMetaValue::from("TIMEOUT"))
        );
    }
    proxy.resume();
    // The timed-out script may have run; a new read must still be safe.
    let recovered = cache.read_by_id(&f.context, 1).await.unwrap();
    f.track(&recovered.stamp);
    proxy.disconnect();
    // The first operation may observe the broken connection or reconnect already.
    let mut read = cache.read_by_id(&f.context, 1).await;
    for _ in 0..10 {
        if read.is_ok() {
            break;
        }
        tokio::time::sleep(Duration::from_millis(20)).await;
        read = cache.read_by_id(&f.context, 1).await;
    }
    let read = read.unwrap();
    f.track(&read.stamp);
    assert!(read.value.is_none());
    f.cleanup().await;
}

#[tokio::test]
#[ignore = "requires isolated Redis via LOCKMATE_TEST_REDIS_URL"]
async fn revision_change_between_snapshot_and_get_cannot_serve_old_entry() {
    let mut f = Fixture::new().await;
    let spaces: Arc<dyn cache::Space> = Arc::new(RedisSpace::new(f.backend.clone()));
    let original = spaces.read_by_id(&f.context, 1).await.unwrap();
    f.track(&original.stamp);
    spaces
        .store_record(
            &f.context,
            &original.stamp,
            &support::space(),
            Duration::from_secs(10),
        )
        .await
        .unwrap();
    let proxy = support::Proxy::new(f.config.redis_host.clone(), f.config.redis_port).await;
    f.config.redis_port = proxy.port;
    f.config.redis_operation_timeout = Duration::from_secs(2);
    let backend = RedisBackend::connect(&f.context, &f.config, f.keys.clone(), f.logger.clone())
        .await
        .unwrap();
    let guarded: Arc<dyn cache::Space> = Arc::new(RedisSpace::new(backend));
    proxy.pause();
    let context = f.context.clone();
    let reading = guarded.clone();
    let read = tokio::spawn(async move { reading.read_by_id(&context, 1).await });
    // The snapshot script has executed, but its response is held by this proxy.
    proxy.wait_held().await;
    RedisInvalidation::new(f.backend.clone())
        .invalidate(&f.context, &[CacheFamily::Spaces])
        .await
        .unwrap();
    proxy.resume();
    let missed = read.await.unwrap().unwrap();
    f.track(&missed.stamp);
    assert!(missed.value.is_none());
    assert!(missed.stamp == original.stamp);
    assert_eq!(
        guarded
            .store_record(
                &f.context,
                &missed.stamp,
                &support::space(),
                Duration::from_secs(1)
            )
            .await
            .unwrap(),
        CacheStoreOutcome::Superseded
    );
    f.cleanup().await;
}
