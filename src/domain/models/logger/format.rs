use std::{fmt, str::FromStr};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum LoggerFormat {
    Plain,
    Json,
}

impl fmt::Display for LoggerFormat {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(match self {
            Self::Plain => "plain",
            Self::Json => "json",
        })
    }
}

impl FromStr for LoggerFormat {
    type Err = &'static str;

    fn from_str(value: &str) -> Result<Self, Self::Err> {
        match value {
            "plain" => Ok(Self::Plain),
            "json" => Ok(Self::Json),
            _ => Err("invalid logger format"),
        }
    }
}
