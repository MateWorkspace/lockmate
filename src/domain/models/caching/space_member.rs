use serde::{Deserialize, Serialize};

use super::CachedUser;
use crate::domain::models::SpaceMember;

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CachedSpaceMemberWithUser {
    pub member: SpaceMember,
    pub user: CachedUser,
}
