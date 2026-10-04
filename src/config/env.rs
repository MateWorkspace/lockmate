use std::{collections::HashMap, env, time::Duration};

use crate::domain::models::{AppEnv, LoggerFormat, LoggerLevel, ValidatorError};

pub fn load_env() -> Result<AppEnv, ValidatorError> {
    let dotenv = read_dotenv(dotenvy::dotenv_iter())?;
    load_env_with(|key| match env::var(key) {
        Ok(value) => Ok(Some(value)),
        Err(env::VarError::NotPresent) => Ok(dotenv.get(key).cloned()),
        Err(env::VarError::NotUnicode(_)) => Err(ValidatorError::EnvironmentInvalid {
            key,
            reason: "must contain valid UTF-8",
        }),
    })
}

fn read_dotenv(
    result: dotenvy::Result<impl Iterator<Item = dotenvy::Result<(String, String)>>>,
) -> Result<HashMap<String, String>, ValidatorError> {
    let entries = match result {
        Ok(entries) => entries,
        Err(error) if error.not_found() => return Ok(HashMap::new()),
        Err(_) => {
            return Err(ValidatorError::EnvironmentInvalid {
                key: ".env",
                reason: "could not be read",
            });
        }
    };
    let mut values = HashMap::new();
    for entry in entries {
        let (key, value) = entry.map_err(|_| ValidatorError::EnvironmentInvalid {
            key: ".env",
            reason: "contains an invalid entry or could not be read",
        })?;
        values.entry(key).or_insert(value);
    }
    Ok(values)
}

fn load_env_with(
    get: impl Fn(&'static str) -> Result<Option<String>, ValidatorError>,
) -> Result<AppEnv, ValidatorError> {
    let reader = EnvReader { get };
    let config = AppEnv {
        logger_format: reader
            .string("LOCKMATE_LOGGER_FORMAT", "json")?
            .trim()
            .to_ascii_lowercase()
            .parse()
            .unwrap_or(LoggerFormat::Json),
        logger_level: reader
            .string("LOCKMATE_LOGGER_LEVEL", "info")?
            .trim()
            .to_ascii_uppercase()
            .parse()
            .unwrap_or(LoggerLevel::Info),

        port_grpc_server: reader.port("LOCKMATE_PORT_GRPC_SERVER", 50051)?,
        port_http_server: reader.port("LOCKMATE_PORT_HTTP_SERVER", 50050)?,

        token_access_secret: reader.required("LOCKMATE_TOKEN_ACCESS_SECRET")?,
        token_refresh_secret: reader.required("LOCKMATE_TOKEN_REFRESH_SECRET")?,
        token_access_duration: reader.duration("LOCKMATE_TOKEN_ACCESS_DURATION", 15 * 60)?,
        token_refresh_duration: reader.duration("LOCKMATE_TOKEN_REFRESH_DURATION", 24 * 60 * 60)?,

        postgres_host: reader.string("LOCKMATE_POSTGRES_HOST", "127.0.0.1")?,
        postgres_port: reader.port("LOCKMATE_POSTGRES_PORT", 5432)?,
        postgres_username: reader.string("LOCKMATE_POSTGRES_USERNAME", "postgres")?,
        postgres_password: reader.string("LOCKMATE_POSTGRES_PASSWORD", "postgres")?,
        postgres_database: reader.string("LOCKMATE_POSTGRES_DATABASE", "lockmate")?,
        postgres_ssl_mode: reader.string("LOCKMATE_POSTGRES_SSL_MODE", "disable")?,
        postgres_max_connections: reader.max_connections()?,
        postgres_connect_timeout: reader.duration("LOCKMATE_POSTGRES_CONNECT_TIMEOUT", 10)?,

        redis_host: reader.string("LOCKMATE_REDIS_HOST", "127.0.0.1")?,
        redis_port: reader.port("LOCKMATE_REDIS_PORT", 6379)?,
        redis_database: reader.redis_database()?,
        redis_password: reader.string("LOCKMATE_REDIS_PASSWORD", "")?,
        redis_namespace: reader.string("LOCKMATE_REDIS_NAMESPACE", "lockmate")?,
        redis_connect_timeout: reader.duration("LOCKMATE_REDIS_CONNECT_TIMEOUT", 5)?,
        redis_operation_timeout: reader.duration_with_fallback(
            "LOCKMATE_REDIS_OPERATION_TIMEOUT",
            Duration::from_millis(200),
        )?,

        caching_record_ttl: reader.duration("LOCKMATE_CACHING_RECORD_TTL", 60)?,
        caching_list_ttl: reader.duration("LOCKMATE_CACHING_LIST_TTL", 15)?,
    };
    Ok(config)
}

struct EnvReader<F> {
    get: F,
}

impl<F: Fn(&'static str) -> Result<Option<String>, ValidatorError>> EnvReader<F> {
    fn string(&self, key: &'static str, fallback: &str) -> Result<String, ValidatorError> {
        Ok((self.get)(key)?
            .filter(|value| !value.is_empty())
            .unwrap_or_else(|| fallback.to_owned()))
    }

    fn required(&self, key: &'static str) -> Result<String, ValidatorError> {
        let value = self.string(key, "")?;
        if value.trim().is_empty() {
            return Err(ValidatorError::EnvironmentRequired { key });
        }
        Ok(value)
    }

    fn integer(&self, key: &'static str, fallback: i64) -> Result<i64, ValidatorError> {
        Ok(self.string(key, "")?.parse().unwrap_or(fallback))
    }

    fn port(&self, key: &'static str, fallback: u16) -> Result<u16, ValidatorError> {
        let value = self.integer(key, i64::from(fallback))?;
        if !(1..=65535).contains(&value) {
            return Err(ValidatorError::EnvironmentInvalid {
                key,
                reason: "must be between 1 and 65535",
            });
        }
        Ok(value as u16)
    }

    fn max_connections(&self) -> Result<u32, ValidatorError> {
        const KEY: &str = "LOCKMATE_POSTGRES_MAX_CONNECTIONS";
        let value = self.integer(KEY, 20)?;
        if !(0..=i64::from(i32::MAX)).contains(&value) {
            return Err(ValidatorError::EnvironmentInvalid {
                key: KEY,
                reason: "must be between 0 and 2147483647",
            });
        }
        Ok(value as u32)
    }

    fn redis_database(&self) -> Result<u8, ValidatorError> {
        const KEY: &str = "LOCKMATE_REDIS_DATABASE";
        let value = self.integer(KEY, 0)?;
        if !(0..=i64::from(u8::MAX)).contains(&value) {
            return Err(ValidatorError::EnvironmentInvalid {
                key: KEY,
                reason: "must be between 0 and 255",
            });
        }
        Ok(value as u8)
    }

    fn duration(&self, key: &'static str, fallback: u64) -> Result<Duration, ValidatorError> {
        self.duration_with_fallback(key, Duration::from_secs(fallback))
    }

    fn duration_with_fallback(
        &self,
        key: &'static str,
        fallback: Duration,
    ) -> Result<Duration, ValidatorError> {
        let value = self.string(key, "")?;
        let negative = value.starts_with('-');
        let unsigned = value.strip_prefix(['-', '+']).unwrap_or(&value);
        let parsed = humantime::parse_duration(unsigned)
            .ok()
            .or_else(|| unsigned.parse::<u64>().ok().map(Duration::from_secs));
        let duration = parsed.unwrap_or(fallback);

        // Malformed values use the fallback; valid negative and zero values are invalid.
        if (negative && parsed.is_some()) || duration.is_zero() {
            return Err(ValidatorError::EnvironmentInvalid {
                key,
                reason: "must be positive",
            });
        }
        Ok(duration)
    }
}
