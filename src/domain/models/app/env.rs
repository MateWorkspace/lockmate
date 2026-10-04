use std::time::Duration;

use crate::domain::models::logger::{LoggerFormat, LoggerLevel};

pub struct AppEnv {
    pub logger_format: LoggerFormat,
    pub logger_level: LoggerLevel,

    pub port_grpc_server: u16,
    pub port_http_server: u16,

    pub token_access_secret: String,
    pub token_refresh_secret: String,
    pub token_access_duration: Duration,
    pub token_refresh_duration: Duration,

    pub postgres_host: String,
    pub postgres_port: u16,
    pub postgres_username: String,
    pub postgres_password: String,
    pub postgres_database: String,
    pub postgres_ssl_mode: String,
    pub postgres_max_connections: u32,
    pub postgres_connect_timeout: Duration,

    pub redis_host: String,
    pub redis_port: u16,
    pub redis_database: u8,
    pub redis_password: String,
    pub redis_namespace: String,
    pub redis_connect_timeout: Duration,
    pub redis_operation_timeout: Duration,

    pub caching_record_ttl: Duration,
    pub caching_list_ttl: Duration,
}
