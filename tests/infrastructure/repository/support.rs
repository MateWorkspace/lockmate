use lockmate::{
    domain::{
        contracts::{
            repository as repo,
            utility::{Logger, Transactor},
        },
        models::{AppContext, LoggerLevel, LoggerMeta},
    },
    infrastructure::{repository::*, utility::transactor::MatePgdtTransactor},
};
use mate_pgdt::{
    Pgdt,
    sqlx::{self, postgres::PgPoolOptions},
};
use std::{
    sync::{
        Arc, Mutex,
        atomic::{AtomicU64, Ordering},
    },
    time::{Duration, SystemTime, UNIX_EPOCH},
};
use uuid::Uuid;

#[derive(Clone)]
pub struct LogEntry {
    pub context: AppContext,
    pub level: LoggerLevel,
    pub tag: String,
    pub message: String,
    pub meta: LoggerMeta,
}

#[derive(Default)]
pub struct RecordingLogger(pub Mutex<Vec<LogEntry>>);

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

pub struct Fixture {
    pub db: Pgdt,
    pub context: AppContext,
    pub logger: Arc<RecordingLogger>,
    pub permissions: Arc<dyn repo::Permission>,
    pub roles: Arc<dyn repo::Role>,
    pub users: Arc<dyn repo::User>,
    pub api_keys: Arc<dyn repo::ApiKey>,
    pub role_permissions: Arc<dyn repo::RolePermission>,
    pub transactor: Arc<dyn Transactor>,
    schema: String,
    url: String,
}

impl Fixture {
    pub async fn new() -> Self {
        static NEXT: AtomicU64 = AtomicU64::new(0);
        let url = std::env::var("LOCKMATE_TEST_DATABASE_URL")
            .expect("set LOCKMATE_TEST_DATABASE_URL to an isolated PostgreSQL database");
        let schema = format!(
            "repo_{}_{}_{}",
            std::process::id(),
            SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_nanos(),
            NEXT.fetch_add(1, Ordering::Relaxed)
        );
        let bootstrap = PgPoolOptions::new()
            .max_connections(1)
            .connect(&url)
            .await
            .unwrap();
        sqlx::query(sqlx::AssertSqlSafe(format!("CREATE SCHEMA {schema}")))
            .execute(&bootstrap)
            .await
            .unwrap();
        bootstrap.close().await;
        let search_path = schema.clone();
        let pool = PgPoolOptions::new()
            .max_connections(4)
            .acquire_timeout(Duration::from_secs(3))
            .after_connect(move |connection, _| {
                let sql = format!("SET search_path TO {search_path}, public");
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
        // The fixture applies the real up migrations, including the avatar column.
        for migration in [
            include_str!("../../../database/migrations/20260930033915_create_permissions.up.sql"),
            include_str!("../../../database/migrations/20260930033922_create_roles.up.sql"),
            include_str!(
                "../../../database/migrations/20260930033930_create_role_permission.up.sql"
            ),
            include_str!("../../../database/migrations/20260930033933_create_users.up.sql"),
            include_str!("../../../database/migrations/20260930033938_create_api_keys.up.sql"),
        ] {
            sqlx::raw_sql(sqlx::AssertSqlSafe(migration))
                .execute(&pool)
                .await
                .unwrap();
        }
        let db = Pgdt::new(pool);
        let logger = Arc::new(RecordingLogger::default());
        Self {
            permissions: Arc::new(PostgresPermission::new(db.clone(), logger.clone())),
            roles: Arc::new(PostgresRole::new(db.clone(), logger.clone())),
            users: Arc::new(PostgresUser::new(db.clone(), logger.clone())),
            api_keys: Arc::new(PostgresApiKey::new(db.clone(), logger.clone())),
            role_permissions: Arc::new(PostgresRolePermission::new(db.clone(), logger.clone())),
            transactor: Arc::new(MatePgdtTransactor::new(db.transactor(), logger.clone())),
            context: AppContext {
                actor: Some("repository tester".into()),
                trace_id: Some(Uuid::from_u128(42)),
                ..AppContext::default()
            },
            db,
            logger,
            schema,
            url,
        }
    }

    pub async fn permission(&self, name: &str) -> i64 {
        self.permissions
            .create(
                &self.context,
                repo::CreatePermission {
                    name: name.into(),
                    description: None,
                    by: Some(42),
                },
            )
            .await
            .unwrap()
    }
    pub async fn role(&self, name: &str, default: bool) -> i64 {
        self.roles
            .create(
                &self.context,
                repo::CreateRole {
                    name: name.into(),
                    description: None,
                    is_default: Some(default),
                    by: Some(42),
                },
            )
            .await
            .unwrap()
    }
    pub fn user_input(role_id: i64, username: &str) -> repo::CreateUser {
        repo::CreateUser {
            role_id,
            name: "User".into(),
            bio: None,
            username: username.into(),
            email: None,
            phone: None,
            password_hash: "private-password-hash".into(),
            is_email_verified: None,
            is_phone_verified: None,
            avatar_path: None,
            by: Some(42),
        }
    }
    pub async fn user(&self, role_id: i64, username: &str) -> i64 {
        self.users
            .create(&self.context, Self::user_input(role_id, username))
            .await
            .unwrap()
    }
    pub fn key_input(user_id: i64, name: &str, hash: &str) -> repo::CreateApiKey {
        repo::CreateApiKey {
            user_id,
            name: name.into(),
            description: None,
            hash: hash.into(),
            redacted: "abcd".into(),
            by: Some(42),
        }
    }
    pub async fn pivot(&self, role_id: i64, permission_id: i64) -> i64 {
        self.role_permissions
            .create(
                &self.context,
                repo::CreateRolePermission {
                    role_id,
                    permission_id,
                    by: Some(42),
                },
            )
            .await
            .unwrap()
    }
    pub async fn close(self) {
        self.db.pool().close().await;
        let pool = PgPoolOptions::new()
            .max_connections(1)
            .connect(&self.url)
            .await
            .unwrap();
        sqlx::query(sqlx::AssertSqlSafe(format!(
            "DROP SCHEMA {} CASCADE",
            self.schema
        )))
        .execute(&pool)
        .await
        .unwrap();
        pool.close().await;
    }
}
