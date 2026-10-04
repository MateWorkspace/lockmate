use crate::domain::models::{
    AppContext, MemberRole as MemberRoleEntity, Permission, Role, SpaceMember,
};

use super::RepositoryFuture;

pub trait MemberRole: Send + Sync {
    fn create<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        input: CreateMemberRole,
    ) -> RepositoryFuture<'a, i64>;

    fn read_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> RepositoryFuture<'a, MemberRoleDetails>;

    fn read_by_member_id_and_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
        role_id: i64,
    ) -> RepositoryFuture<'a, MemberRoleDetails>;

    fn read_by_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        role_id: i64,
    ) -> RepositoryFuture<'a, Vec<MemberRoleWithMember>>;

    fn read_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
    ) -> RepositoryFuture<'a, Vec<MemberRoleWithRole>>;

    fn read_effective_roles_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
    ) -> RepositoryFuture<'a, Vec<Role>>;

    fn read_effective_permissions_by_member_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: i64,
    ) -> RepositoryFuture<'a, Vec<Permission>>;

    fn delete_by_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        id: i64,
    ) -> RepositoryFuture<'a, ()>;

    fn delete_by_member_id_or_role_id<'a>(
        &'a self,
        context: &'a AppContext,
        space_id: i64,
        member_id: Option<i64>,
        role_id: Option<i64>,
    ) -> RepositoryFuture<'a, ()>;
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CreateMemberRole {
    pub member_id: i64,
    pub role_id: i64,
    pub by: Option<i64>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct MemberRoleDetails {
    pub member_role: MemberRoleEntity,
    pub member: SpaceMember,
    pub role: Role,
}

#[derive(Debug, Clone, PartialEq)]
pub struct MemberRoleWithMember {
    pub member_role: MemberRoleEntity,
    pub member: SpaceMember,
}

#[derive(Debug, Clone, PartialEq)]
pub struct MemberRoleWithRole {
    pub member_role: MemberRoleEntity,
    pub role: Role,
}
