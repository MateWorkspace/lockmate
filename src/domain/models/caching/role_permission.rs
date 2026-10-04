use serde::{Deserialize, Serialize};

use crate::domain::models::{Permission, Role, RolePermission};

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CachedRolePermissionDetails {
    pub role_permission: RolePermission,
    pub role: Role,
    pub permission: Permission,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CachedRolePermissionWithPermission {
    pub role_permission: RolePermission,
    pub permission: Permission,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CachedRolePermissionWithRole {
    pub role_permission: RolePermission,
    pub role: Role,
}
