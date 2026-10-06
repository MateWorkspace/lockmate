use serde::{Deserialize, Serialize};

use super::UserView;
use crate::domain::models::SpaceMember;

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct SpaceMemberWithUser {
    pub member: SpaceMember,
    pub user: UserView,
}

impl From<crate::domain::contracts::repository::SpaceMemberWithUser> for SpaceMemberWithUser {
    fn from(value: crate::domain::contracts::repository::SpaceMemberWithUser) -> Self {
        Self {
            member: value.member,
            user: value.user.into(),
        }
    }
}

impl From<crate::domain::models::CachedSpaceMemberWithUser> for SpaceMemberWithUser {
    fn from(value: crate::domain::models::CachedSpaceMemberWithUser) -> Self {
        Self {
            member: value.member,
            user: value.user.into(),
        }
    }
}
