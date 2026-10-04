use lockmate::{
    domain::{
        contracts::{caching::KeyBuilder, utility::Logger},
        models::*,
    },
    infrastructure::caching::*,
};
use redis::{ConnectionAddr, IntoConnectionInfo};
use std::{
    collections::BTreeSet,
    sync::{Arc, Mutex},
    time::Duration,
};

pub struct Entry {
    pub context: AppContext,
    pub level: LoggerLevel,
    pub tag: String,
    pub message: String,
    pub meta: LoggerMeta,
}

#[derive(Default)]
pub struct RecordingLogger(pub Mutex<Vec<Entry>>);
impl Logger for RecordingLogger {
    fn log(
        &self,
        context: &AppContext,
        level: LoggerLevel,
        tag: &str,
        message: &str,
        meta: &LoggerMeta,
    ) {
        self.0.lock().unwrap().push(Entry {
            context: context.clone(),
            level,
            tag: tag.into(),
            message: message.into(),
            meta: meta.clone(),
        });
    }
}

pub fn env() -> AppEnv {
    AppEnv {
        logger_format: LoggerFormat::Json,
        logger_level: LoggerLevel::Info,
        port_grpc_server: 50051,
        port_http_server: 50050,
        token_access_secret: String::new(),
        token_refresh_secret: String::new(),
        token_access_duration: Duration::from_secs(900),
        token_refresh_duration: Duration::from_secs(86400),
        postgres_host: String::new(),
        postgres_port: 5432,
        postgres_username: String::new(),
        postgres_password: String::new(),
        postgres_database: String::new(),
        postgres_ssl_mode: String::new(),
        postgres_max_connections: 20,
        postgres_connect_timeout: Duration::from_secs(10),
        redis_host: "127.0.0.1".into(),
        redis_port: 6379,
        redis_database: 0,
        redis_password: String::new(),
        redis_namespace: "test".into(),
        redis_connect_timeout: Duration::from_secs(5),
        redis_operation_timeout: Duration::from_millis(200),
        caching_record_ttl: Duration::from_secs(60),
        caching_list_ttl: Duration::from_secs(15),
    }
}

pub struct Fixture {
    pub backend: RedisBackend,
    pub raw: redis::aio::MultiplexedConnection,
    pub keys: Arc<Sha256KeyBuilder>,
    pub logger: Arc<RecordingLogger>,
    pub context: AppContext,
    pub config: AppEnv,
    tracked: Mutex<BTreeSet<String>>,
}

impl Fixture {
    pub async fn new() -> Self {
        let url = std::env::var("LOCKMATE_TEST_REDIS_URL")
            .expect("set LOCKMATE_TEST_REDIS_URL to isolated Redis");
        let info = url.into_connection_info().unwrap();
        let mut config = env();
        let ConnectionAddr::Tcp(host, port) = info.addr() else {
            panic!("tests require standalone TCP Redis")
        };
        config.redis_host = host.clone();
        config.redis_port = *port;
        config.redis_password = info.redis_settings().password().unwrap_or("").into();
        config.redis_database = u8::try_from(info.redis_settings().db()).unwrap();
        let mut bytes = [0; 16];
        getrandom::fill(&mut bytes).unwrap();
        config.redis_namespace = format!("cache_tests_{}", hex::encode(bytes));
        let keys = Arc::new(Sha256KeyBuilder::new(&config.redis_namespace).unwrap());
        let logger = Arc::new(RecordingLogger::default());
        let context = AppContext {
            actor: Some("cache test".into()),
            trace_id: Some(uuid::Uuid::from_u128(1)),
            transaction: None,
        };
        let backend = RedisBackend::connect(&context, &config, keys.clone(), logger.clone())
            .await
            .unwrap();
        let raw = redis::Client::open(info)
            .unwrap()
            .get_multiplexed_async_connection()
            .await
            .unwrap();
        Self {
            backend,
            raw,
            keys,
            logger,
            context,
            config,
            tracked: Mutex::new(BTreeSet::new()),
        }
    }

    pub fn track(&self, stamp: &CacheStamp) -> String {
        let entry = self.keys.entry(&self.context, &stamp.key).unwrap();
        self.tracked.lock().unwrap().insert(entry.clone());
        for revision in &stamp.key.revisions {
            self.revision(revision.family);
        }
        entry
    }

    pub fn revision(&self, family: CacheFamily) -> String {
        let key = self.keys.revision(&self.context, family).unwrap();
        self.tracked.lock().unwrap().insert(key.clone());
        key
    }

    pub async fn cleanup(self) {
        let keys: Vec<_> = self.tracked.lock().unwrap().iter().cloned().collect();
        if !keys.is_empty() {
            redis::cmd("DEL")
                .arg(keys)
                .query_async::<i64>(&mut self.raw.clone())
                .await
                .unwrap();
        }
    }
}

pub fn audit() -> AuditCreateUpdateDelete {
    AuditCreateUpdateDelete {
        create: AuditCreate {
            at: time::OffsetDateTime::UNIX_EPOCH,
            by: None,
        },
        update: AuditUpdate::default(),
        delete: AuditDelete::default(),
    }
}
pub fn space() -> Space {
    Space {
        id: 1,
        slug: "example".into(),
        name: "Example".into(),
        description: String::new(),
        is_active: true,
        preferences: serde_json::json!({}),
        audit: audit(),
    }
}
pub fn user() -> CachedUser {
    CachedUser {
        id: 2,
        name: "Example".into(),
        bio: String::new(),
        username: "example".into(),
        email: Some("example@example.com".into()),
        phone: Some("1234567890".into()),
        is_email_verified: false,
        is_phone_verified: false,
        avatar_path: None,
        preferences: serde_json::json!({}),
        audit: audit(),
    }
}
pub fn permission() -> Permission {
    Permission {
        id: 3,
        space_id: 1,
        slug: "example.read".into(),
        name: "Example Read".into(),
        description: String::new(),
        preferences: serde_json::json!({}),
        audit: audit(),
    }
}
pub fn role() -> Role {
    Role {
        id: 4,
        space_id: 1,
        slug: "example".into(),
        name: "Example".into(),
        description: String::new(),
        is_default: true,
        preferences: serde_json::json!({}),
        audit: audit(),
    }
}
pub fn member() -> SpaceMember {
    SpaceMember {
        id: 5,
        space_id: 1,
        user_id: 2,
        is_active: true,
        preferences: serde_json::json!({}),
        audit: audit(),
    }
}
pub fn api_key() -> CachedApiKey {
    CachedApiKey {
        id: 6,
        space_id: 1,
        member_id: 5,
        name: "Example key".into(),
        description: String::new(),
        redacted: "abcd".into(),
        preferences: serde_json::json!({}),
        audit: audit(),
    }
}
pub fn role_permission() -> RolePermission {
    RolePermission {
        id: 7,
        space_id: 1,
        role_id: 4,
        permission_id: 3,
        audit: audit().create,
    }
}
pub fn member_role() -> MemberRole {
    MemberRole {
        id: 8,
        space_id: 1,
        member_id: 5,
        role_id: 4,
        audit: audit().create,
    }
}

/// Per-test TCP relay: pause responses without pausing a shared Redis server.
pub struct Proxy {
    pub port: u16,
    blocked: Arc<std::sync::atomic::AtomicBool>,
    release: Arc<tokio::sync::Notify>,
    held: Arc<tokio::sync::Notify>,
    workers: Arc<Mutex<Vec<tokio::task::AbortHandle>>>,
    accept: tokio::task::JoinHandle<()>,
}

impl Proxy {
    pub async fn new(host: String, port: u16) -> Self {
        use tokio::io::{AsyncReadExt, AsyncWriteExt};
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let proxy_port = listener.local_addr().unwrap().port();
        let blocked = Arc::new(std::sync::atomic::AtomicBool::new(false));
        let release = Arc::new(tokio::sync::Notify::new());
        let held = Arc::new(tokio::sync::Notify::new());
        let workers = Arc::new(Mutex::new(Vec::new()));
        let hold = blocked.clone();
        let signal = release.clone();
        let handles = workers.clone();
        let waiting = held.clone();
        let accept = tokio::spawn(async move {
            loop {
                let (incoming, _) = listener.accept().await.unwrap();
                let upstream = tokio::net::TcpStream::connect((host.as_str(), port))
                    .await
                    .unwrap();
                let (mut client_read, mut client_write) = incoming.into_split();
                let (mut upstream_read, mut upstream_write) = upstream.into_split();
                let hold = hold.clone();
                let signal = signal.clone();
                let waiting = waiting.clone();
                let worker = tokio::spawn(async move {
                    let requests = tokio::io::copy(&mut client_read, &mut upstream_write);
                    let responses = async {
                        let mut bytes = [0; 8192];
                        loop {
                            let count = upstream_read.read(&mut bytes).await?;
                            if count == 0 {
                                return Ok::<(), std::io::Error>(());
                            }
                            // Register before checking to avoid losing a release notification.
                            let released = signal.notified();
                            tokio::pin!(released);
                            released.as_mut().enable();
                            if hold.load(std::sync::atomic::Ordering::SeqCst) && bytes[0] == b'*' {
                                waiting.notify_one();
                                released.await;
                            }
                            client_write.write_all(&bytes[..count]).await?;
                        }
                    };
                    tokio::select! { _ = requests => {}, _ = responses => {} }
                });
                handles.lock().unwrap().push(worker.abort_handle());
            }
        });
        Self {
            port: proxy_port,
            blocked,
            release,
            held,
            workers,
            accept,
        }
    }

    pub async fn wait_held(&self) {
        tokio::time::timeout(Duration::from_secs(1), self.held.notified())
            .await
            .unwrap();
    }

    pub fn pause(&self) {
        self.blocked
            .store(true, std::sync::atomic::Ordering::SeqCst);
    }
    pub fn resume(&self) {
        self.blocked
            .store(false, std::sync::atomic::Ordering::SeqCst);
        self.release.notify_waiters();
    }
    pub fn disconnect(&self) {
        for worker in self.workers.lock().unwrap().drain(..) {
            worker.abort();
        }
    }
}

impl Drop for Proxy {
    fn drop(&mut self) {
        self.accept.abort();
        self.disconnect();
    }
}
