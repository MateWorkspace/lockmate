use std::{
    io::{self, Write},
    sync::{Arc, Mutex},
    thread,
};

use lockmate::{
    domain::{
        contracts::utility::Logger,
        models::{
            AppContext, LoggerFormat, LoggerLevel, LoggerMeta, LoggerMetaValue, RepositoryError,
        },
    },
    infrastructure::utility::logger::{JsonLogger, PlainLogger},
};
use serde_json::{Value, json};
use uuid::Uuid;

#[derive(Clone, Default)]
struct Capture(Arc<Mutex<Vec<u8>>>);

impl Write for Capture {
    fn write(&mut self, bytes: &[u8]) -> io::Result<usize> {
        self.0.lock().unwrap().extend_from_slice(bytes);
        Ok(bytes.len())
    }

    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}

impl Capture {
    fn output(&self) -> String {
        String::from_utf8(self.0.lock().unwrap().clone()).unwrap()
    }
}

fn new_logger(writer: Capture, format: LoggerFormat, level: LoggerLevel) -> Box<dyn Logger> {
    match format {
        LoggerFormat::Json => Box::new(JsonLogger::with_writer(writer, level)),
        LoggerFormat::Plain => Box::new(PlainLogger::with_writer(writer, level)),
    }
}

#[test]
fn absent_context_fields_and_empty_metadata_have_explicit_defaults() {
    for format in [LoggerFormat::Json, LoggerFormat::Plain] {
        let capture = Capture::default();
        let logger = new_logger(capture.clone(), format, LoggerLevel::Info);
        logger.info(&AppContext::default(), "", "event", &LoggerMeta::new());
        let output = capture.output();
        assert_eq!(output.lines().count(), 1);
        match format {
            LoggerFormat::Json => {
                let record: Value = serde_json::from_str(output.trim()).unwrap();
                assert_eq!(record["tag"], "");
                assert_eq!(record["actor"], "UNKNOWN");
                assert_eq!(record["trace_id"], Uuid::nil().to_string());
                assert_eq!(record["meta"], json!({}));
                assert_eq!(record["message"], "event");
                assert_eq!(record["level"], "INFO");
                assert!(record["time"].as_str().unwrap().ends_with('Z'));
            }
            LoggerFormat::Plain => {
                assert!(output.contains(" INFO tag=\"\" actor=\"UNKNOWN\""));
                assert!(output.contains("trace_id=00000000-0000-0000-0000-000000000000"));
                assert!(output.ends_with("message=\"event\" meta={}\n"));
            }
        }
    }
}

#[test]
fn prints_supplied_context_and_metadata_without_expanding_error_sources() {
    let capture = Capture::default();
    let logger = JsonLogger::with_writer(capture.clone(), LoggerLevel::Error);
    let context = AppContext {
        actor: Some("admin".into()),
        trace_id: Some(Uuid::from_u128(42)),
        ..AppContext::default()
    };
    let error = RepositoryError::Failure {
        source: Box::new(io::Error::other("private database details")),
    };
    let meta = LoggerMeta::from([
        ("id".into(), 42_i64.into()),
        ("error".into(), LoggerMetaValue::from_error(&error)),
        (
            "details".into(),
            LoggerMetaValue::Object(LoggerMeta::from([(
                "flags".into(),
                LoggerMetaValue::Array(vec![true.into(), LoggerMetaValue::Null]),
            )])),
        ),
    ]);
    logger.error(&context, "repository/user/ReadById", "read failed", &meta);
    let output = capture.output();
    let record: Value = serde_json::from_str(output.trim()).unwrap();
    assert_eq!(record["tag"], "repository/user/ReadById");
    assert_eq!(record["actor"], "admin");
    assert_eq!(record["trace_id"], Uuid::from_u128(42).to_string());
    assert_eq!(
        record["meta"],
        json!({
            "id": 42,
            "error": "repository operation failed",
            "details": {"flags": [true, null]},
        })
    );
    assert!(!output.contains("private database details"));
}

#[test]
fn plain_output_preserves_empty_actor_and_escapes_control_characters() {
    let capture = Capture::default();
    let logger = PlainLogger::with_writer(capture.clone(), LoggerLevel::Warn);
    logger.warn(
        &AppContext {
            actor: Some(String::new()),
            trace_id: None,
            ..AppContext::default()
        },
        "repository\n/read",
        "failed\n\u{1b}",
        &LoggerMeta::from([("attempt".into(), 2_u64.into())]),
    );
    let output = capture.output();
    assert_eq!(output.lines().count(), 1);
    assert!(output.contains("tag=\"repository\\n/read\" actor=\"\""));
    assert!(output.contains("message=\"failed\\n\\u001b\" meta={\"attempt\":2}"));
}

#[test]
fn filters_all_levels_in_both_formats() {
    let levels = [
        LoggerLevel::None,
        LoggerLevel::Error,
        LoggerLevel::Warn,
        LoggerLevel::Info,
        LoggerLevel::Debug,
    ];
    for format in [LoggerFormat::Json, LoggerFormat::Plain] {
        for (count, configured) in levels.into_iter().enumerate() {
            let capture = Capture::default();
            let logger = new_logger(capture.clone(), format, configured);
            for level in levels {
                logger.log(
                    &AppContext::default(),
                    level,
                    "test/log",
                    "event",
                    &LoggerMeta::new(),
                );
            }
            assert_eq!(capture.output().lines().count(), count);
        }
    }
}

#[test]
fn shared_logger_keeps_concurrent_records_and_context_intact() {
    let capture = Capture::default();
    let logger: Arc<dyn Logger> =
        Arc::new(JsonLogger::with_writer(capture.clone(), LoggerLevel::Info));
    let threads: Vec<_> = (1..=4)
        .map(|id| {
            let logger = Arc::clone(&logger);
            thread::spawn(move || {
                let context = AppContext {
                    actor: Some(id.to_string()),
                    trace_id: Some(Uuid::from_u128(id)),
                    ..AppContext::default()
                };
                for _ in 0..20 {
                    logger.info(
                        &context,
                        "repository/read",
                        &id.to_string(),
                        &LoggerMeta::new(),
                    );
                }
            })
        })
        .collect();
    for handle in threads {
        handle.join().unwrap();
    }
    let output = capture.output();
    assert_eq!(output.lines().count(), 80);
    for line in output.lines() {
        let record: Value = serde_json::from_str(line).unwrap();
        let id = record["message"].as_str().unwrap().parse::<u128>().unwrap();
        assert_eq!(record["actor"], id.to_string());
        assert_eq!(record["trace_id"], Uuid::from_u128(id).to_string());
    }
}
