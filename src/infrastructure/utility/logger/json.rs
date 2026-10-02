use std::{
    io::{self, Write},
    sync::Mutex,
};

use crate::domain::{
    contracts::utility::Logger,
    models::{LoggerContext, LoggerLevel, LoggerMeta},
};

use super::utils::{LogRecord, write_record};

pub struct JsonLogger<W = io::Stderr> {
    writer: Mutex<W>,
    level: LoggerLevel,
}

impl JsonLogger {
    pub fn new(level: LoggerLevel) -> Self {
        Self::with_writer(io::stderr(), level)
    }
}

impl<W: Write + Send> JsonLogger<W> {
    pub fn with_writer(writer: W, level: LoggerLevel) -> Self {
        Self {
            writer: Mutex::new(writer),
            level,
        }
    }
}

impl<W: Write + Send> Logger for JsonLogger<W> {
    fn log(&self, context: &LoggerContext, level: LoggerLevel, message: &str, meta: &LoggerMeta) {
        if !self.level.allows(level) {
            return;
        }

        let record = LogRecord::new(context, level, message, meta);
        let Ok(record) = serde_json::to_string(&record) else {
            return;
        };

        write_record(&self.writer, &record);
    }
}
