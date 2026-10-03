use crate::domain::models::{AppContext, LoggerLevel, LoggerMeta};

pub trait Logger: Send + Sync {
    fn log(
        &self,
        context: &AppContext,
        level: LoggerLevel,
        tag: &str,
        message: &str,
        meta: &LoggerMeta,
    );

    fn error(&self, context: &AppContext, tag: &str, message: &str, meta: &LoggerMeta) {
        self.log(context, LoggerLevel::Error, tag, message, meta);
    }

    fn warn(&self, context: &AppContext, tag: &str, message: &str, meta: &LoggerMeta) {
        self.log(context, LoggerLevel::Warn, tag, message, meta);
    }

    fn info(&self, context: &AppContext, tag: &str, message: &str, meta: &LoggerMeta) {
        self.log(context, LoggerLevel::Info, tag, message, meta);
    }

    fn debug(&self, context: &AppContext, tag: &str, message: &str, meta: &LoggerMeta) {
        self.log(context, LoggerLevel::Debug, tag, message, meta);
    }
}
