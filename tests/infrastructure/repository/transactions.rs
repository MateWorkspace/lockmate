use super::support::{Fixture, RecordingLogger};
use lockmate::{
    domain::{
        contracts::repository::{CreatePermission, Permission, PermissionFilter, UpdateRole},
        models::{RepositoryError, TransactorError},
    },
    infrastructure::repository::PostgresPermission,
};
use mate_pgdt::{
    Pgdt,
    sqlx::{self, Row, postgres::PgPoolOptions},
};
use std::{error::Error as _, sync::Arc};
use tokio::sync::oneshot;

#[tokio::test]
async fn repository_trait_objects_return_send_futures_without_connecting() {
    fn send<T: Send>(_: T) {}
    let db = Pgdt::new(
        PgPoolOptions::new()
            .connect_lazy("postgres://localhost/unused")
            .unwrap(),
    );
    let repository: Arc<dyn Permission> = Arc::new(PostgresPermission::new(
        db.clone(),
        Arc::new(RecordingLogger::default()),
    ));
    let context = Default::default();
    send(repository.read_by_slug(&context, 1, "borrowed"));
    let logger = Arc::new(RecordingLogger::default());
    use lockmate::{domain::contracts::repository as repo, infrastructure::repository::*};
    let spaces: Arc<dyn repo::Space> = Arc::new(PostgresSpace::new(db.clone(), logger.clone()));
    let users: Arc<dyn repo::User> = Arc::new(PostgresUser::new(db.clone(), logger.clone()));
    let roles: Arc<dyn repo::Role> = Arc::new(PostgresRole::new(db.clone(), logger.clone()));
    let members: Arc<dyn repo::SpaceMember> =
        Arc::new(PostgresSpaceMember::new(db.clone(), logger.clone()));
    let assignments: Arc<dyn repo::MemberRole> =
        Arc::new(PostgresMemberRole::new(db.clone(), logger.clone()));
    let grants: Arc<dyn repo::RolePermission> =
        Arc::new(PostgresRolePermission::new(db.clone(), logger.clone()));
    let keys: Arc<dyn repo::ApiKey> = Arc::new(PostgresApiKey::new(db.clone(), logger));
    send(spaces.read_by_slug(&context, "borrowed"));
    send(users.read_by_username(&context, "borrowed"));
    send(roles.read_default(&context, 1));
    send(members.read_by_user_id(&context, 1, 1));
    send(assignments.read_effective_permissions_by_member_id(&context, 1, 1));
    send(grants.read_by_role_id_and_permission_id(&context, 1, 1, 1));
    send(keys.read_active_by_hash(&context, 1, "borrowed"));
    db.pool().close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn cross_repository_transactions_commit_rollback_and_reject_retained_handles() {
    let f = Fixture::new().await;
    let space_id = f.space_id;
    let role = f.role("default", true).await;
    let permissions = f.permissions.clone();
    let roles = f.roles.clone();
    let db = &f.db;
    let root = &f.context;
    let (send, receive) = oneshot::channel();
    f.transactor
        .with_tx(
            root,
            Box::new(move |context| {
                Box::pin(async move {
                    permissions
                        .create(
                            &context,
                            space_id,
                            CreatePermission {
                                slug: "committed".into(),
                                name: "committed".into(),
                                description: None,
                                by: None,
                            },
                        )
                        .await?;
                    roles
                        .update_by_id(
                            &context,
                            space_id,
                            role,
                            UpdateRole {
                                is_default: Some(true),
                                ..Default::default()
                            },
                        )
                        .await?;
                    let count: i64 = db
                        .query_row(
                            root,
                            sqlx::query("SELECT COUNT(*) AS count FROM permissions"),
                        )
                        .await
                        .map_err(|source| RepositoryError::Failure {
                            source: Box::new(source),
                        })?
                        .try_get("count")
                        .unwrap();
                    assert_eq!(count, 0);
                    send.send(context).unwrap();
                    Ok(())
                })
            }),
        )
        .await
        .unwrap();
    let retained = receive.await.unwrap();
    let error = f
        .permissions
        .read_by_id(&retained, f.space_id, 1)
        .await
        .unwrap_err();
    assert!(matches!(error, RepositoryError::Failure { .. }));
    assert!(matches!(
        error.source().unwrap().downcast_ref::<sqlx::Error>(),
        Some(sqlx::Error::Protocol(_))
    ));
    let permissions = f.permissions.clone();
    let roles = f.roles.clone();
    let result = f
        .transactor
        .with_tx(
            root,
            Box::new(move |context| {
                Box::pin(async move {
                    permissions
                        .create(
                            &context,
                            space_id,
                            CreatePermission {
                                slug: "rolled-back".into(),
                                name: "rolled-back".into(),
                                description: None,
                                by: None,
                            },
                        )
                        .await?;
                    roles
                        .update_by_id(
                            &context,
                            space_id,
                            role,
                            UpdateRole {
                                name: Some("rolled-back".into()),
                                ..Default::default()
                            },
                        )
                        .await?;
                    Err(RepositoryError::BadState.into())
                })
            }),
        )
        .await;
    assert!(matches!(result, Err(TransactorError::Callback { .. })));
    assert_eq!(
        f.roles
            .read_by_id(root, f.space_id, role)
            .await
            .unwrap()
            .name,
        "default"
    );
    assert!(matches!(
        f.permissions
            .read_by_slug(root, f.space_id, "rolled-back")
            .await,
        Err(RepositoryError::PermissionNotFound)
    ));
    assert!(
        f.permissions
            .read_by_slug(root, f.space_id, "committed")
            .await
            .is_ok()
    );
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn foreign_handles_failure_sources_and_safe_logging() {
    let f = Fixture::new().await;
    let space_id = f.space_id;
    let id = f.permission("private-permission-name").await;
    let foreign = PostgresPermission::new(Pgdt::new(f.db.pool().clone()), f.logger.clone());
    let result = f
        .transactor
        .with_tx(
            &f.context,
            Box::new(move |context| {
                Box::pin(async move {
                    let error = foreign
                        .read_by_id(&context, space_id, id)
                        .await
                        .unwrap_err();
                    assert!(matches!(error, RepositoryError::Failure { .. }));
                    Ok(())
                })
            }),
        )
        .await;
    result.unwrap();
    let entries = f.logger.0.lock().unwrap().clone();
    assert_eq!(entries.len(), 1);
    assert_eq!(entries[0].tag, "repository/permission/postgres/read_by_id");
    f.logger.0.lock().unwrap().clear();
    assert!(matches!(
        f.permissions
            .create(
                &f.context,
                f.space_id,
                CreatePermission {
                    slug: "private-permission-name".into(),
                    name: "private-permission-name".into(),
                    description: None,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::PermissionSlugConflict)
    ));
    let entries = f.logger.0.lock().unwrap().clone();
    assert_eq!(entries.len(), 1);
    assert_eq!(entries[0].meta["sql_state"], "23505".into());
    assert_eq!(
        entries[0].meta["constraint"],
        "uq_permissions_space_id_slug".into()
    );
    assert_eq!(entries[0].context, f.context);
    assert!(!entries[0].message.contains("private-permission-name"));
    f.logger.0.lock().unwrap().clear();
    f.db.exec(
        &f.context,
        sqlx::query(
            "ALTER TABLE permissions ADD CONSTRAINT invalid_name CHECK (name <> 'rejected')",
        ),
    )
    .await
    .unwrap();
    assert!(matches!(
        f.permissions
            .create(
                &f.context,
                f.space_id,
                CreatePermission {
                    slug: "rejected".into(),
                    name: "rejected".into(),
                    description: None,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    // Changing a selected column's type exercises fallible decoding and its source.
    f.db.exec(
        &f.context,
        sqlx::query(
            "ALTER TABLE permissions ALTER COLUMN created_by TYPE TEXT USING created_by::TEXT",
        ),
    )
    .await
    .unwrap();
    let error = f
        .permissions
        .read_by_id(&f.context, f.space_id, id)
        .await
        .unwrap_err();
    assert!(matches!(error, RepositoryError::Failure { .. }));
    assert!(
        error
            .source()
            .unwrap()
            .downcast_ref::<sqlx::Error>()
            .is_some()
    );
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn closed_pool_errors_log_once_and_keep_native_cause() {
    let f = Fixture::new().await;
    f.db.pool().close().await;
    let error = f
        .permissions
        .read_by_filter(
            &f.context,
            f.space_id,
            PermissionFilter {
                page: 1,
                limit: 10,
                search: None,
            },
        )
        .await
        .unwrap_err();
    assert!(matches!(error, RepositoryError::Failure { .. }));
    assert!(matches!(
        error.source().unwrap().downcast_ref::<sqlx::Error>(),
        Some(sqlx::Error::PoolClosed)
    ));
    assert_eq!(f.logger.0.lock().unwrap().len(), 1);
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn statement_timeout_and_rollback_preserve_single_repository_log() {
    let f = Fixture::new().await;
    let space_id = f.space_id;
    let permissions = f.permissions.clone();
    let db = &f.db;
    let result = f
        .transactor
        .with_tx(
            &f.context,
            Box::new(move |context| {
                Box::pin(async move {
                    db.exec(
                        &context,
                        sqlx::query("SET LOCAL statement_timeout = '50ms'"),
                    )
                    .await
                    .map_err(|source| RepositoryError::Failure {
                        source: Box::new(source),
                    })?;
                    // A trigger causes a native query cancellation without logging its private inputs.
                    permissions
                        .create(
                            &context,
                            space_id,
                            CreatePermission {
                                slug: "triggered".into(),
                                name: "triggered".into(),
                                description: None,
                                by: None,
                            },
                        )
                        .await?;
                    Ok(())
                })
            }),
        )
        .await;
    // Without a trigger this succeeds; cancellation behavior is tested below via a trigger.
    result.unwrap();
    f.db.exec(&f.context,sqlx::query("CREATE FUNCTION slow_permission() RETURNS TRIGGER LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_sleep(0.2); RETURN NEW; END $$")).await.unwrap();
    f.db.exec(&f.context,sqlx::query("CREATE TRIGGER slow_permission BEFORE INSERT ON permissions FOR EACH ROW EXECUTE FUNCTION slow_permission()")).await.unwrap();
    let permissions = f.permissions.clone();
    let result = f
        .transactor
        .with_tx(
            &f.context,
            Box::new(move |context| {
                Box::pin(async move {
                    db.exec(
                        &context,
                        sqlx::query("SET LOCAL statement_timeout = '50ms'"),
                    )
                    .await
                    .map_err(|source| RepositoryError::Failure {
                        source: Box::new(source),
                    })?;
                    permissions
                        .create(
                            &context,
                            space_id,
                            CreatePermission {
                                slug: "private-input".into(),
                                name: "private-input".into(),
                                description: None,
                                by: None,
                            },
                        )
                        .await?;
                    Ok(())
                })
            }),
        )
        .await;
    assert!(matches!(result, Err(TransactorError::Callback { .. })));
    let entries = f.logger.0.lock().unwrap().clone();
    assert_eq!(entries.len(), 1);
    assert_eq!(entries[0].meta["sql_state"], "57014".into());
    assert!(!entries[0].message.contains("private-input"));
    let (_, total) = f
        .permissions
        .read_by_filter(
            &f.context,
            f.space_id,
            PermissionFilter {
                page: 1,
                limit: 10,
                search: None,
            },
        )
        .await
        .unwrap();
    assert_eq!(total, 1);
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn pool_timeouts_are_classified_and_preserve_native_source() {
    let f = Fixture::new().await;
    let mut connections = Vec::new();
    for _ in 0..4 {
        connections.push(f.db.pool().acquire().await.unwrap());
    }
    let error = f
        .permissions
        .read_by_id(&f.context, f.space_id, 1)
        .await
        .unwrap_err();
    assert!(matches!(error, RepositoryError::Timeout { .. }));
    assert!(matches!(
        error.source().unwrap().downcast_ref::<sqlx::Error>(),
        Some(sqlx::Error::PoolTimedOut)
    ));
    let entries = f.logger.0.lock().unwrap().clone();
    assert_eq!(entries.len(), 1);
    assert_eq!(entries[0].meta["error_code"], "TIMEOUT".into());
    drop(connections);
    f.close().await;
}
