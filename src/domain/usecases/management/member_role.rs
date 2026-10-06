use crate::domain::models::{AccessClaims, AppContext};
use crate::domain::usecases::{
    UsecaseFuture,
    shared::{MemberRoleDetails, MemberRoleWithMember, MemberRoleWithRole},
};

pub trait MemberRole: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        input: CreateRequest,
    ) -> UsecaseFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        id: i64,
    ) -> UsecaseFuture<'a, MemberRoleDetails>;

    fn read_by_member_id_and_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        member_id: i64,
        role_id: i64,
    ) -> UsecaseFuture<'a, MemberRoleDetails>;

    fn read_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        role_id: i64,
    ) -> UsecaseFuture<'a, Vec<MemberRoleWithMember>>;

    fn read_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        member_id: i64,
    ) -> UsecaseFuture<'a, Vec<MemberRoleWithRole>>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        id: i64,
    ) -> UsecaseFuture<'a, ()>;

    fn delete_by_member_id_or_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        space_id: i64,
        member_id: Option<i64>,
        role_id: Option<i64>,
    ) -> UsecaseFuture<'a, ()>;
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CreateRequest {
    pub member_id: i64,
    pub role_id: i64,
}
