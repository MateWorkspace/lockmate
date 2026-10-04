use serde::{Deserialize, Serialize};

/// Page contents and total from one repository result, cached together.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CachePage<T> {
    pub items: Vec<T>,
    pub total: i64,
}

impl<T> From<(Vec<T>, i64)> for CachePage<T> {
    fn from((items, total): (Vec<T>, i64)) -> Self {
        Self { items, total }
    }
}
