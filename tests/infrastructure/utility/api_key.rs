use std::sync::{Arc, Mutex};

use lockmate::{
    domain::{
        contracts::utility::{ApiKey, Logger},
        models::{ApiKeyError, AppContext, LoggerLevel, LoggerMeta, LoggerMetaValue},
    },
    infrastructure::utility::api_key::ApiKeyGenerator,
};

#[derive(Default)]
struct RecordingLogger {
    records: Mutex<Vec<(AppContext, String, String, LoggerMeta)>>,
}

impl Logger for RecordingLogger {
    fn log(
        &self,
        context: &AppContext,
        level: LoggerLevel,
        tag: &str,
        message: &str,
        meta: &LoggerMeta,
    ) {
        assert_eq!(level, LoggerLevel::Error);
        self.records.lock().unwrap().push((
            context.clone(),
            tag.to_owned(),
            message.to_owned(),
            meta.clone(),
        ));
    }
}

#[test]
fn generates_fixed_length_keys_and_matching_storage_values() {
    let logger = Arc::new(RecordingLogger::default());
    let generator: Arc<dyn ApiKey> = Arc::new(ApiKeyGenerator::new(logger.clone()));
    let context = AppContext::default();
    let first = generator.generate(&context).unwrap();
    let second = generator.generate(&context).unwrap();
    assert_ne!(first.raw, second.raw);
    for key in [first, second] {
        assert_eq!(key.raw.len(), 73);
        assert!(key.raw.starts_with("lockmate-"));
        generator.validate(&context, &key.raw).unwrap();
        assert_eq!(key.hash, generator.hash(&context, &key.raw));
        assert_eq!(key.hash.len(), 64);
        assert_eq!(key.redacted, key.raw[key.raw.len() - 4..]);
    }
    assert!(logger.records.lock().unwrap().is_empty());
}

#[test]
fn hashes_like_go_sha256_and_accepts_leading_zero_bytes() {
    let generator = ApiKeyGenerator::new(Arc::new(RecordingLogger::default()));
    let context = AppContext::default();
    assert_eq!(
        generator.hash(&context, "abc"),
        "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
    );
    let raw = format!("lockmate-{}01", "0".repeat(62));
    generator.validate(&context, &raw).unwrap();
    assert_eq!(generator.hash(&context, &raw).len(), 64);
}

#[test]
fn rejects_wrong_prefix_length_alphabet_and_whitespace() {
    let generator = ApiKeyGenerator::new(Arc::new(RecordingLogger::default()));
    let context = AppContext::default();
    for raw in [
        format!("nadi_{}", "a".repeat(64)),
        format!("Lockmate-{}", "a".repeat(64)),
        format!("lockmate-{}", "a".repeat(63)),
        format!("lockmate-{}", "a".repeat(65)),
        format!("lockmate-{}", "A".repeat(64)),
        format!("lockmate-{}g", "a".repeat(63)),
        format!("lockmate-{}é", "a".repeat(62)),
        format!(" lockmate-{}", "a".repeat(64)),
        format!("lockmate-{}\n", "a".repeat(64)),
    ] {
        assert!(matches!(
            generator.validate(&context, &raw),
            Err(ApiKeyError::Invalid)
        ));
    }
    assert!(matches!(
        generator.validate(&context, ""),
        Err(ApiKeyError::BadArgs)
    ));
}

#[test]
fn validation_errors_keep_context_and_log_once_without_keys() {
    let logger = Arc::new(RecordingLogger::default());
    let generator = ApiKeyGenerator::new(logger.clone());
    let context = AppContext {
        actor: Some("test actor".to_owned()),
        trace_id: Some(uuid::Uuid::from_u128(123)),
        ..AppContext::default()
    };
    assert!(generator.validate(&context, "").is_err());
    let secret = "lockmate-invalid-private-key";
    assert!(generator.validate(&context, secret).is_err());

    let records = logger.records.lock().unwrap();
    assert_eq!(records.len(), 2);
    for (record, code) in records.iter().zip(["BAD_ARGS", "API_KEY_INVALID"]) {
        let (logged_context, tag, message, meta) = record;
        assert_eq!(logged_context, &context);
        assert!(tag.starts_with("utility/api_key/generator/"));
        assert_eq!(meta.get("error_code"), Some(&LoggerMetaValue::from(code)));
        assert!(!message.contains(secret));
        assert!(!format!("{meta:?}").contains(secret));
    }
}
