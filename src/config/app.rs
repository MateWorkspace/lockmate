use crate::domain::models::app::AppInfo;

pub const APP_INFO: AppInfo = AppInfo {
    name: env!("CARGO_PKG_NAME"),
    version: env!("CARGO_PKG_VERSION"),
};
