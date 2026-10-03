use std::{
    error::Error as _,
    future::pending,
    sync::{
        Arc, Mutex,
        atomic::{AtomicU64, Ordering},
    },
    time::{Duration, SystemTime, UNIX_EPOCH},
};

use lockmate::{
    domain::{
        contracts::utility::{Logger, Transactor},
        models::{
            AppContext, AppTransaction, LoggerLevel, LoggerMeta, RepositoryError, TransactorError,
        },
    },
    infrastructure::utility::transactor::MatePgdtTransactor,
};
use mate_pgdt::{
    Pgdt,
    sqlx::{self, PgPool, Row, postgres::PgPoolOptions},
};
use tokio::sync::oneshot;
use uuid::Uuid;

struct LogEntry {
    context: AppContext,
    level: LoggerLevel,
    tag: String,
    message: String,
    meta: LoggerMeta,
}

#[derive(Default)]
struct RecordingLogger(Mutex<Vec<LogEntry>>);

impl Logger for RecordingLogger {
    fn log(
        &self,
        context: &AppContext,
        level: LoggerLevel,
        tag: &str,
        message: &str,
        meta: &LoggerMeta,
    ) {
        self.0.lock().unwrap().push(LogEntry {
            context: context.clone(),
            level,
            tag: tag.into(),
            message: message.into(),
            meta: meta.clone(),
        });
    }
}

struct Fixture {
    db: Pgdt,
    transactor: Arc<dyn Transactor>,
    logger: Arc<RecordingLogger>,
}

fn context() -> AppContext {
    AppContext {
        actor: Some("transaction tester".into()),
        trace_id: Some(Uuid::from_u128(42)),
        ..AppContext::default()
    }
}

fn adapter(db: &Pgdt, logger: Arc<RecordingLogger>) -> Arc<dyn Transactor> {
    Arc::new(MatePgdtTransactor::new(db.transactor(), logger))
}

async fn fixture(max_connections: u32) -> Fixture {
    static NEXT: AtomicU64 = AtomicU64::new(0);
    let url = std::env::var("LOCKMATE_TEST_DATABASE_URL")
        .expect("set LOCKMATE_TEST_DATABASE_URL to an isolated PostgreSQL database");
    let schema = format!(
        "lockmate_tx_{}_{}_{}",
        std::process::id(),
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_nanos(),
        NEXT.fetch_add(1, Ordering::Relaxed)
    );
    let bootstrap = PgPool::connect(&url).await.unwrap();
    sqlx::query(sqlx::AssertSqlSafe(format!("CREATE SCHEMA {schema}")))
        .execute(&bootstrap)
        .await
        .unwrap();
    bootstrap.close().await;
    let pool = PgPoolOptions::new()
        .max_connections(max_connections)
        .acquire_timeout(Duration::from_secs(2))
        .after_connect(move |connection, _| {
            let sql = format!("SET search_path TO {schema}");
            Box::pin(async move {
                sqlx::query(sqlx::AssertSqlSafe(sql))
                    .execute(connection)
                    .await?;
                Ok(())
            })
        })
        .connect(&url)
        .await
        .unwrap();
    let db = Pgdt::new(pool);
    db.exec(
        &AppContext::default(),
        sqlx::query("CREATE TABLE items (id BIGINT PRIMARY KEY)"),
    )
    .await
    .unwrap();
    let logger = Arc::new(RecordingLogger::default());
    let transactor = adapter(&db, logger.clone());
    Fixture {
        db,
        transactor,
        logger,
    }
}

fn repository_error(source: sqlx::Error) -> RepositoryError {
    RepositoryError::Failure {
        source: Box::new(source),
    }
}

async fn insert(db: &Pgdt, context: &AppContext, id: i64) -> Result<(), RepositoryError> {
    db.exec(
        context,
        sqlx::query("INSERT INTO items VALUES ($1)").bind(id),
    )
    .await
    .map_err(repository_error)?;
    Ok(())
}

async fn count(db: &Pgdt) -> i64 {
    db.query_row(
        &AppContext::default(),
        sqlx::query("SELECT COUNT(*) AS count FROM items"),
    )
    .await
    .unwrap()
    .try_get("count")
    .unwrap()
}

fn driver_source(error: &TransactorError) -> &sqlx::Error {
    let mut source = error.source().unwrap();
    loop {
        if let Some(driver) = source.downcast_ref::<sqlx::Error>() {
            return driver;
        }
        source = source.source().unwrap();
    }
}

fn callback_source(error: &TransactorError) -> &RepositoryError {
    assert!(matches!(error, TransactorError::Callback { .. }));
    error
        .source()
        .unwrap()
        .downcast_ref::<RepositoryError>()
        .unwrap()
}

fn assert_log(logger: &RecordingLogger, context: &AppContext, code: &str) {
    let entries = logger.0.lock().unwrap();
    assert_eq!(entries.len(), 1);
    let entry = &entries[0];
    assert_eq!(&entry.context, context);
    assert_eq!(entry.level, LoggerLevel::Error);
    assert_eq!(entry.tag, "utility/transactor/mate_pgdt/with_tx");
    assert_eq!(entry.meta["error_code"], code.into());
    assert_eq!(
        entry.message,
        if code == "TIMEOUT" {
            "transaction operation timed out"
        } else {
            "transaction operation failed"
        }
    );
}

#[tokio::test]
async fn object_safe_contract_has_send_futures_and_borrowed_callbacks() {
    fn assert_send<T: Send>(_: T) {}
    let db = Pgdt::new(
        PgPoolOptions::new()
            .connect_lazy("postgres://localhost/unused")
            .unwrap(),
    );
    let transactor = adapter(&db, Arc::new(RecordingLogger::default()));
    let context = context();
    let borrowed = String::from("borrowed");
    let borrowed = &borrowed;
    let original = &context;
    assert_send(transactor.with_tx(
        &context,
        Box::new(|derived| {
            Box::pin(async move {
                assert_eq!(derived.actor, original.actor);
                assert_eq!(borrowed, "borrowed");
                Ok(())
            })
        }),
    ));
    db.pool().close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn commits_nested_work_preserves_metadata_and_invalidates_handles() {
    let fixture = fixture(2).await;
    let db = &fixture.db;
    let transactor = &fixture.transactor;
    let context = context();
    let original = &context;
    let (send, receive) = oneshot::channel();
    transactor
        .with_tx(
            original,
            Box::new(move |outer| {
                Box::pin(async move {
                    assert_eq!(outer.actor, original.actor);
                    assert_eq!(outer.trace_id, original.trace_id);
                    insert(db, &outer, 1).await?;
                    assert_eq!(count(db).await, 0);
                    let handle = outer.transaction.clone();
                    let result = transactor
                        .with_tx(
                            &outer,
                            Box::new(move |inner| {
                                Box::pin(async move {
                                    assert_eq!(inner.transaction, handle);
                                    insert(db, &inner, 2).await?;
                                    Err(RepositoryError::Conflict.into())
                                })
                            }),
                        )
                        .await;
                    assert!(matches!(
                        callback_source(&result.unwrap_err()),
                        RepositoryError::Conflict
                    ));
                    send.send(outer).unwrap();
                    Ok(())
                })
            }),
        )
        .await
        .unwrap();
    assert_eq!(count(db).await, 2);
    assert!(context.transaction.is_none());
    assert!(fixture.logger.0.lock().unwrap().is_empty());
    let retained = receive.await.unwrap();
    assert!(
        !mate_pgdt::TransactionContext::transaction(&retained)
            .unwrap()
            .is_active()
    );
    let error = transactor
        .with_tx(&retained, Box::new(|_| Box::pin(async { Ok(()) })))
        .await
        .unwrap_err();
    assert!(matches!(driver_source(&error), sqlx::Error::Protocol(_)));
    assert_log(&fixture.logger, &retained, "FAILURE");
    db.pool().close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn callback_errors_roll_back_unchanged_without_adapter_logs() {
    let fixture = fixture(1).await;
    let db = &fixture.db;
    let context = context();
    let result = fixture
        .transactor
        .with_tx(
            &context,
            Box::new(move |context| {
                Box::pin(async move {
                    insert(db, &context, 1).await?;
                    Err(RepositoryError::UserEmailConflict.into())
                })
            }),
        )
        .await;
    assert!(matches!(
        callback_source(&result.unwrap_err()),
        RepositoryError::UserEmailConflict
    ));
    assert_eq!(count(db).await, 0);
    let error = fixture
        .transactor
        .with_tx(
            &context,
            Box::new(move |context| {
                Box::pin(async move {
                    insert(db, &context, 2).await?;
                    insert(db, &context, 2).await.map_err(Into::into)
                })
            }),
        )
        .await
        .unwrap_err();
    assert_eq!(
        driver_source(&error)
            .as_database_error()
            .unwrap()
            .code()
            .as_deref(),
        Some("23505")
    );
    assert_eq!(count(db).await, 0);
    assert!(fixture.logger.0.lock().unwrap().is_empty());
    db.pool().close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn begin_timeout_and_closed_pool_keep_native_sources_and_log_once() {
    let fixture = fixture(1).await;
    let context = context();
    let held = fixture.db.pool().acquire().await.unwrap();
    let error = fixture
        .transactor
        .with_tx(
            &context,
            Box::new(|_| Box::pin(async { panic!("callback must not run") })),
        )
        .await
        .unwrap_err();
    assert!(matches!(error, TransactorError::Timeout { .. }));
    assert!(matches!(driver_source(&error), sqlx::Error::PoolTimedOut));
    assert_log(&fixture.logger, &context, "TIMEOUT");
    drop(held);
    fixture.logger.0.lock().unwrap().clear();
    fixture.db.pool().close().await;
    let error = fixture
        .transactor
        .with_tx(&context, Box::new(|_| Box::pin(async { Ok(()) })))
        .await
        .unwrap_err();
    assert!(matches!(driver_source(&error), sqlx::Error::PoolClosed));
    assert_log(&fixture.logger, &context, "FAILURE");
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn independently_wrapped_pool_is_rejected_without_running_callback() {
    let fixture = fixture(1).await;
    let foreign_db = Pgdt::new(fixture.db.pool().clone());
    let foreign = adapter(&foreign_db, fixture.logger.clone());
    let context = context();
    fixture
        .transactor
        .with_tx(
            &context,
            Box::new(move |derived| {
                Box::pin(async move {
                    let error = foreign
                        .with_tx(
                            &derived,
                            Box::new(|_| Box::pin(async { panic!("foreign callback") })),
                        )
                        .await
                        .unwrap_err();
                    assert!(matches!(driver_source(&error), sqlx::Error::Protocol(_)));
                    Ok(())
                })
            }),
        )
        .await
        .unwrap();
    assert_eq!(fixture.logger.0.lock().unwrap().len(), 1);
    fixture.db.pool().close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn commit_failure_keeps_sqlstate_and_constraint_and_logs_safe_message() {
    let fixture = fixture(1).await;
    let db = &fixture.db;
    let context = context();
    db.exec(&context, sqlx::query("CREATE TABLE deferred_items (id BIGINT CONSTRAINT deferred_unique UNIQUE DEFERRABLE INITIALLY DEFERRED)"))
        .await.unwrap();
    let (send, receive) = oneshot::channel();
    let error = fixture
        .transactor
        .with_tx(
            &context,
            Box::new(move |derived| {
                Box::pin(async move {
                    db.exec(
                        &derived,
                        sqlx::query("INSERT INTO deferred_items VALUES (1), (1)"),
                    )
                    .await
                    .map_err(repository_error)?;
                    send.send(derived).unwrap();
                    Ok(())
                })
            }),
        )
        .await
        .unwrap_err();
    assert_eq!(
        driver_source(&error)
            .as_database_error()
            .unwrap()
            .code()
            .as_deref(),
        Some("23505")
    );
    assert!(
        !mate_pgdt::TransactionContext::transaction(&receive.await.unwrap())
            .unwrap()
            .is_active()
    );
    assert_log(&fixture.logger, &context, "FAILURE");
    {
        let entries = fixture.logger.0.lock().unwrap();
        assert_eq!(entries[0].meta["sql_state"], "23505".into());
        assert_eq!(entries[0].meta["constraint"], "deferred_unique".into());
    }
    db.pool().close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn panic_and_cancellation_roll_back_with_escaped_contexts() {
    let fixture = fixture(1).await;
    for panic in [true, false] {
        let db = fixture.db.clone();
        let transactor = fixture.transactor.clone();
        let (send, receive) = oneshot::channel();
        let task = tokio::spawn(async move {
            transactor
                .with_tx(
                    &context(),
                    Box::new(move |context| {
                        Box::pin(async move {
                            insert(&db, &context, 1).await?;
                            send.send(context).unwrap();
                            if panic {
                                panic!("transaction panic");
                            }
                            pending::<()>().await;
                            Ok(())
                        })
                    }),
                )
                .await
        });
        let retained = tokio::time::timeout(Duration::from_secs(5), receive)
            .await
            .unwrap()
            .unwrap();
        if !panic {
            task.abort();
        }
        assert_eq!(task.await.unwrap_err().is_panic(), panic);
        assert!(
            !mate_pgdt::TransactionContext::transaction(&retained)
                .unwrap()
                .is_active()
        );
        assert_eq!(count(&fixture.db).await, 0);
    }
    assert!(fixture.logger.0.lock().unwrap().is_empty());
    fixture.db.pool().close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn secondary_rollback_failure_preserves_callback_error_without_logging() {
    let fixture = fixture(2).await;
    let db = &fixture.db;
    let original = context();
    let root = &original;
    let result = fixture
        .transactor
        .with_tx(
            root,
            Box::new(move |derived| {
                Box::pin(async move {
                    let pid: i32 = db
                        .query_row(&derived, sqlx::query("SELECT pg_backend_pid() AS pid"))
                        .await
                        .map_err(repository_error)?
                        .try_get("pid")
                        .map_err(repository_error)?;
                    db.query_row(
                        root,
                        sqlx::query("SELECT pg_terminate_backend($1)").bind(pid),
                    )
                    .await
                    .map_err(repository_error)?;
                    Err(RepositoryError::UserNotFound.into())
                })
            }),
        )
        .await;
    assert!(matches!(
        callback_source(&result.unwrap_err()),
        RepositoryError::UserNotFound
    ));
    assert!(fixture.logger.0.lock().unwrap().is_empty());
    db.pool().close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn cancellation_waits_for_in_flight_query_then_rolls_back() {
    let fixture = fixture(2).await;
    let db = fixture.db.clone();
    let transactor = fixture.transactor.clone();
    let (send, receive) = oneshot::channel();
    let owner = tokio::spawn(async move {
        transactor
            .with_tx(
                &context(),
                Box::new(move |derived| {
                    Box::pin(async move {
                        insert(&db, &derived, 1).await?;
                        let query_context = derived.clone();
                        let query = tokio::spawn(async move {
                            db.exec(&query_context, sqlx::query("SELECT pg_sleep(1)"))
                                .await
                        });
                        send.send((derived, query)).unwrap();
                        pending::<()>().await;
                        Ok(())
                    })
                }),
            )
            .await
    });
    let (retained, query) = receive.await.unwrap();
    tokio::time::timeout(Duration::from_secs(5), async {
        loop {
            let active: i64 = fixture.db.query_row(&AppContext::default(), sqlx::query(
                "SELECT COUNT(*) AS count FROM pg_stat_activity WHERE query = 'SELECT pg_sleep(1)' AND state = 'active'"
            )).await.unwrap().try_get("count").unwrap();
            if active > 0 { break; }
            tokio::time::sleep(Duration::from_millis(10)).await;
        }
    }).await.unwrap();
    owner.abort();
    assert!(owner.await.unwrap_err().is_cancelled());
    assert!(
        !mate_pgdt::TransactionContext::transaction(&retained)
            .unwrap()
            .is_active()
    );
    query.await.unwrap().unwrap();
    tokio::time::timeout(Duration::from_secs(5), async {
        while count(&fixture.db).await != 0 {
            tokio::task::yield_now().await;
        }
    })
    .await
    .unwrap();
    assert!(fixture.logger.0.lock().unwrap().is_empty());
    fixture.db.pool().close().await;
}

#[test]
fn opaque_handle_preserves_identity_and_hides_contents() {
    let handle = AppTransaction::new(String::from("private transaction data"));
    assert_eq!(
        handle.downcast_ref::<String>().unwrap(),
        "private transaction data"
    );
    assert!(handle.downcast_ref::<u64>().is_none());
    assert_eq!(handle, handle.clone());
    assert_ne!(
        handle,
        AppTransaction::new(String::from("private transaction data"))
    );
    assert!(!format!("{handle:?}").contains("private transaction data"));
}

#[test]
#[should_panic(expected = "AppContext transaction must contain a mate-pgdt handle")]
fn incompatible_handle_cannot_fall_back_to_non_transaction_execution() {
    let context = AppContext {
        transaction: Some(AppTransaction::new(42_u64)),
        ..AppContext::default()
    };
    let _ = mate_pgdt::TransactionContext::transaction(&context);
}

#[test]
fn transactor_errors_classify_failures_and_preserve_original_causes() {
    let callback: TransactorError = RepositoryError::UserEmailConflict.into();
    assert_eq!(callback.code(), "FAILURE");
    assert!(matches!(
        callback_source(&callback),
        RepositoryError::UserEmailConflict
    ));
    let failure = TransactorError::Failure {
        source: Box::new(std::io::Error::other("private driver details")),
    };
    assert_eq!(failure.code(), "FAILURE");
    assert_eq!(failure.to_string(), "transaction operation failed");
    assert!(
        failure
            .source()
            .unwrap()
            .downcast_ref::<std::io::Error>()
            .is_some()
    );
    let timeout = TransactorError::Timeout {
        source: Box::new(std::io::Error::new(
            std::io::ErrorKind::TimedOut,
            "private timeout details",
        )),
    };
    assert_eq!(timeout.code(), "TIMEOUT");
    assert_eq!(timeout.to_string(), "transaction operation timed out");
    let callback = TransactorError::Callback {
        source: Box::new(std::io::Error::other("application callback error")),
    };
    assert!(
        callback
            .source()
            .unwrap()
            .downcast_ref::<std::io::Error>()
            .is_some()
    );
}
