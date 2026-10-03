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
    send(repository.read_by_name(&context, "borrowed"));
    db.pool().close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn cross_repository_transactions_commit_rollback_and_reject_retained_handles() {
    let f = Fixture::new().await;
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
                            CreatePermission {
                                name: "committed".into(),
                                description: None,
                                by: None,
                            },
                        )
                        .await?;
                    roles
                        .update_by_id(
                            &context,
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
    let error = f.permissions.read_by_id(&retained, 1).await.unwrap_err();
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
                            CreatePermission {
                                name: "rolled-back".into(),
                                description: None,
                                by: None,
                            },
                        )
                        .await?;
                    roles
                        .update_by_id(
                            &context,
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
        f.roles.read_by_id(root, role).await.unwrap().name,
        "default"
    );
    assert!(matches!(
        f.permissions.read_by_name(root, "rolled-back").await,
        Err(RepositoryError::PermissionNotFound)
    ));
    assert!(f.permissions.read_by_name(root, "committed").await.is_ok());
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn foreign_handles_failure_sources_and_safe_logging() {
    let f = Fixture::new().await;
    let id = f.permission("private-permission-name").await;
    let foreign = PostgresPermission::new(Pgdt::new(f.db.pool().clone()), f.logger.clone());
    let result = f
        .transactor
        .with_tx(
            &f.context,
            Box::new(move |context| {
                Box::pin(async move {
                    let error = foreign.read_by_id(&context, id).await.unwrap_err();
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
                CreatePermission {
                    name: "private-permission-name".into(),
                    description: None,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::PermissionNameConflict)
    ));
    let entries = f.logger.0.lock().unwrap().clone();
    assert_eq!(entries.len(), 1);
    assert_eq!(entries[0].meta["sql_state"], "23505".into());
    assert_eq!(entries[0].meta["constraint"], "uq_permissions_name".into());
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
                CreatePermission {
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
    let error = f.permissions.read_by_id(&f.context, id).await.unwrap_err();
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
                            CreatePermission {
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
                            CreatePermission {
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
    let error = f.permissions.read_by_id(&f.context, 1).await.unwrap_err();
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
