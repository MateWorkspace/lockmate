use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Page<T> {
    pub items: Vec<T>,
    pub total: i64,
}

impl<T> From<(Vec<T>, i64)> for Page<T> {
    fn from((items, total): (Vec<T>, i64)) -> Self {
        Self { items, total }
    }
}

impl<T> From<crate::domain::models::CachePage<T>> for Page<T> {
    fn from(value: crate::domain::models::CachePage<T>) -> Self {
        Self {
            items: value.items,
            total: value.total,
        }
    }
}
