use std::{io::Write, sync::Mutex};

use serde::Serialize;
use time::{OffsetDateTime, format_description::well_known::Rfc3339};

use crate::domain::models::{LoggerContext, LoggerLevel, LoggerMeta};

#[derive(Serialize)]
pub(super) struct LogRecord<'a> {
    pub time: String,
    pub level: &'static str,
    pub tag: &'a str,
    pub actor: &'a str,
    pub trace_id: String,
    pub message: &'a str,
    pub meta: &'a LoggerMeta,
}

impl<'a> LogRecord<'a> {
    pub fn new(
        context: &'a LoggerContext,
        level: LoggerLevel,
        message: &'a str,
        meta: &'a LoggerMeta,
    ) -> Self {
        let now = OffsetDateTime::now_utc();
        Self {
            time: now
                .format(&Rfc3339)
                .unwrap_or_else(|_| now.unix_timestamp().to_string()),
            level: level.as_str(),
            tag: &context.tag,
            actor: context.actor.as_deref().unwrap_or("UNKNOWN"),
            trace_id: context.trace_id.unwrap_or_default().to_string(),
            message,
            meta,
        }
    }
}

pub(super) fn write_record<W: Write>(writer: &Mutex<W>, record: &str) {
    let mut writer = writer.lock().unwrap_or_else(|error| error.into_inner());
    let _ = writeln!(writer, "{record}");
    let _ = writer.flush();
}
