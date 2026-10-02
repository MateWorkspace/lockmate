use std::{
    io::{self, Write},
    sync::Mutex,
};

use serde_json::json;

use crate::domain::{
    contracts::utility::Logger,
    models::{LoggerContext, LoggerLevel, LoggerMeta},
};

use super::utils::{LogRecord, write_record};

pub struct PlainLogger<W = io::Stderr> {
    writer: Mutex<W>,
    level: LoggerLevel,
}

impl PlainLogger {
    pub fn new(level: LoggerLevel) -> Self {
        Self::with_writer(io::stderr(), level)
    }
}

impl<W: Write + Send> PlainLogger<W> {
    pub fn with_writer(writer: W, level: LoggerLevel) -> Self {
        Self {
            writer: Mutex::new(writer),
            level,
        }
    }
}

impl<W: Write + Send> Logger for PlainLogger<W> {
    fn log(&self, context: &LoggerContext, level: LoggerLevel, message: &str, meta: &LoggerMeta) {
        if !self.level.allows(level) {
            return;
        }

        let record = LogRecord::new(context, level, message, meta);
        let Ok(meta) = serde_json::to_string(record.meta) else {
            return;
        };
        let record = format!(
            "{} {} tag={} actor={} trace_id={} message={} meta={}",
            record.time,
            record.level,
            json!(record.tag),
            json!(record.actor),
            record.trace_id,
            json!(record.message),
            meta,
        );

        write_record(&self.writer, &record);
    }
}
