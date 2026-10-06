use serde::{Deserialize, Serialize};

use crate::domain::models::{Permission, Role, RolePermission};

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct RolePermissionDetails {
    pub role_permission: RolePermission,
    pub role: Role,
    pub permission: Permission,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct RolePermissionWithPermission {
    pub role_permission: RolePermission,
    pub permission: Permission,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct RolePermissionWithRole {
    pub role_permission: RolePermission,
    pub role: Role,
}

impl From<crate::domain::contracts::repository::RolePermissionDetails> for RolePermissionDetails {
    fn from(value: crate::domain::contracts::repository::RolePermissionDetails) -> Self {
        Self {
            role_permission: value.role_permission,
            role: value.role,
            permission: value.permission,
        }
    }
}

impl From<crate::domain::models::CachedRolePermissionDetails> for RolePermissionDetails {
    fn from(value: crate::domain::models::CachedRolePermissionDetails) -> Self {
        Self {
            role_permission: value.role_permission,
            role: value.role,
            permission: value.permission,
        }
    }
}

impl From<crate::domain::contracts::repository::RolePermissionWithPermission>
    for RolePermissionWithPermission
{
    fn from(value: crate::domain::contracts::repository::RolePermissionWithPermission) -> Self {
        Self {
            role_permission: value.role_permission,
            permission: value.permission,
        }
    }
}

impl From<crate::domain::models::CachedRolePermissionWithPermission>
    for RolePermissionWithPermission
{
    fn from(value: crate::domain::models::CachedRolePermissionWithPermission) -> Self {
        Self {
            role_permission: value.role_permission,
            permission: value.permission,
        }
    }
}

impl From<crate::domain::contracts::repository::RolePermissionWithRole> for RolePermissionWithRole {
    fn from(value: crate::domain::contracts::repository::RolePermissionWithRole) -> Self {
        Self {
            role_permission: value.role_permission,
            role: value.role,
        }
    }
}

impl From<crate::domain::models::CachedRolePermissionWithRole> for RolePermissionWithRole {
    fn from(value: crate::domain::models::CachedRolePermissionWithRole) -> Self {
        Self {
            role_permission: value.role_permission,
            role: value.role,
        }
    }
}
