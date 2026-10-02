use crate::domain::models::{LoggerContext, LoggerLevel, LoggerMeta};

/// Thread-safe, best-effort logging. Context and metadata are supplied by the caller.
pub trait Logger: Send + Sync {
    fn log(&self, context: &LoggerContext, level: LoggerLevel, message: &str, meta: &LoggerMeta);

    fn error(&self, context: &LoggerContext, message: &str, meta: &LoggerMeta) {
        self.log(context, LoggerLevel::Error, message, meta);
    }

    fn warn(&self, context: &LoggerContext, message: &str, meta: &LoggerMeta) {
        self.log(context, LoggerLevel::Warn, message, meta);
    }

    fn info(&self, context: &LoggerContext, message: &str, meta: &LoggerMeta) {
        self.log(context, LoggerLevel::Info, message, meta);
    }

    fn debug(&self, context: &LoggerContext, message: &str, meta: &LoggerMeta) {
        self.log(context, LoggerLevel::Debug, message, meta);
    }
}
