use std::{fmt, str::FromStr};

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub enum LoggerLevel {
    None,
    Error,
    Warn,
    Info,
    Debug,
}

impl LoggerLevel {
    pub fn allows(self, level: Self) -> bool {
        level != Self::None && self >= level
    }

    pub const fn as_str(self) -> &'static str {
        match self {
            Self::None => "NONE",
            Self::Error => "ERROR",
            Self::Warn => "WARN",
            Self::Info => "INFO",
            Self::Debug => "DEBUG",
        }
    }
}

impl fmt::Display for LoggerLevel {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(self.as_str())
    }
}

impl FromStr for LoggerLevel {
    type Err = &'static str;

    fn from_str(value: &str) -> Result<Self, Self::Err> {
        match value {
            "NONE" => Ok(Self::None),
            "ERROR" => Ok(Self::Error),
            "WARN" => Ok(Self::Warn),
            "INFO" => Ok(Self::Info),
            "DEBUG" => Ok(Self::Debug),
            _ => Err("invalid logger level"),
        }
    }
}
