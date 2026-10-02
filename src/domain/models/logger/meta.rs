use std::{collections::BTreeMap, error::Error};

use serde::Serialize;

pub type LoggerMeta = BTreeMap<String, LoggerMetaValue>;

#[derive(Debug, Clone, PartialEq, Serialize)]
#[serde(untagged)]
pub enum LoggerMetaValue {
    Null,
    Bool(bool),
    Integer(i64),
    Unsigned(u64),
    Float(f64),
    String(String),
    Array(Vec<LoggerMetaValue>),
    Object(LoggerMeta),
}

impl LoggerMetaValue {
    pub fn from_error(error: &dyn Error) -> Self {
        Self::String(error.to_string())
    }
}

impl From<&str> for LoggerMetaValue {
    fn from(value: &str) -> Self {
        Self::String(value.to_owned())
    }
}

impl From<String> for LoggerMetaValue {
    fn from(value: String) -> Self {
        Self::String(value)
    }
}

impl From<bool> for LoggerMetaValue {
    fn from(value: bool) -> Self {
        Self::Bool(value)
    }
}

impl From<i64> for LoggerMetaValue {
    fn from(value: i64) -> Self {
        Self::Integer(value)
    }
}

impl From<u64> for LoggerMetaValue {
    fn from(value: u64) -> Self {
        Self::Unsigned(value)
    }
}

impl From<f64> for LoggerMetaValue {
    fn from(value: f64) -> Self {
        Self::Float(value)
    }
}
