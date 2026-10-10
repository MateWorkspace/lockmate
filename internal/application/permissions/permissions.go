// Package applicationpermissions defines the permissions required by application operations.
package applicationpermissions

// Management classic auth operations.
const (
	ManagementClassicAuthCreate         = "classic_auth:add"
	ManagementClassicAuthReadById       = "classic_auth:get"
	ManagementClassicAuthReadByUserId   = "classic_auth:get"
	ManagementClassicAuthReadByUsername = "classic_auth:get"
	ManagementClassicAuthReadByEmail    = "classic_auth:get"
	ManagementClassicAuthReadByPhone    = "classic_auth:get"
	ManagementClassicAuthReadByFilter   = "classic_auth:get"
	ManagementClassicAuthUpdateById     = "classic_auth:set"
	ManagementClassicAuthResetPassword  = "user_password:set"
	ManagementClassicAuthDeleteById     = "classic_auth:remove"
)

// Management client operations.
const (
	ManagementClientCreate       = "client:add"
	ManagementClientReadById     = "client:get"
	ManagementClientReadBySlug   = "client:get"
	ManagementClientReadByFilter = "client:get"
	ManagementClientUpdateById   = "client:set"
	ManagementClientRotateSecret = "client_secret:set"
	ManagementClientDeleteById   = "client:remove"
)

// Management google auth operations.
const (
	ManagementGoogleAuthReadById      = "google_auth:get"
	ManagementGoogleAuthReadByUserId  = "google_auth:get"
	ManagementGoogleAuthReadBySubject = "google_auth:get"
	ManagementGoogleAuthReadByFilter  = "google_auth:get"
)

// Management permission operations.
const (
	ManagementPermissionCreate       = "permission:add"
	ManagementPermissionReadById     = "permission:get"
	ManagementPermissionReadBySlug   = "permission:get"
	ManagementPermissionReadByFilter = "permission:get"
	ManagementPermissionUpdateById   = "permission:set"
	ManagementPermissionDeleteById   = "permission:remove"
)

// Management role operations.
const (
	ManagementRoleCreate       = "role:add"
	ManagementRoleReadById     = "role:get"
	ManagementRoleReadBySlug   = "role:get"
	ManagementRoleReadByFilter = "role:get"
	ManagementRoleReadDefault  = "role:get"
	ManagementRoleUpdateById   = "role:set"
	ManagementRoleSetDefault   = "role:set"
	ManagementRoleDeleteById   = "role:remove"
)

// Management role permission operations.
const (
	ManagementRolePermissionCreate                       = "role_permission:add"
	ManagementRolePermissionReadById                     = "role_permission:get"
	ManagementRolePermissionReadByRoleIdAndPermissionId  = "role_permission:get"
	ManagementRolePermissionReadByRoleId                 = "role_permission:get"
	ManagementRolePermissionReadByPermissionId           = "role_permission:get"
	ManagementRolePermissionDeleteById                   = "role_permission:remove"
	ManagementRolePermissionDeleteByRoleIdOrPermissionId = "role_permission:remove"
)

// Management space operations.
const (
	ManagementSpaceCreate       = "space:add"
	ManagementSpaceReadById     = "space:get"
	ManagementSpaceReadBySlug   = "space:get"
	ManagementSpaceReadByFilter = "space:get"
	ManagementSpaceUpdateById   = "space:set"
	ManagementSpaceDeleteById   = "space:remove"
)

// Management space member operations.
const (
	ManagementSpaceMemberCreate                                  = "space_member:add"
	ManagementSpaceMemberReadById                                = "space_member:get"
	ManagementSpaceMemberReadByUserId                            = "space_member:get"
	ManagementSpaceMemberReadByFilter                            = "space_member:get"
	ManagementSpaceMemberUpdateById                              = "space_member:set"
	ManagementSpaceMemberReadEffectiveRolesBySpaceMemberId       = "user_permission:get"
	ManagementSpaceMemberReadEffectivePermissionsBySpaceMemberId = "user_permission:get"
	ManagementSpaceMemberDeleteById                              = "space_member:remove"
)

// Management space member role operations.
const (
	ManagementSpaceMemberRoleCreate                        = "space_member_role:add"
	ManagementSpaceMemberRoleReadById                      = "space_member_role:get"
	ManagementSpaceMemberRoleReadBySpaceMemberIdAndRoleId  = "space_member_role:get"
	ManagementSpaceMemberRoleReadBySpaceMemberId           = "space_member_role:get"
	ManagementSpaceMemberRoleReadByRoleId                  = "space_member_role:get"
	ManagementSpaceMemberRoleDeleteById                    = "space_member_role:remove"
	ManagementSpaceMemberRoleDeleteBySpaceMemberIdOrRoleId = "space_member_role:remove"
)

// Management user operations.
const (
	ManagementUserCreate       = "user:add"
	ManagementUserReadById     = "user:get"
	ManagementUserReadByFilter = "user:get"
	ManagementUserUpdateById   = "user:set"
	ManagementUserDeleteById   = "user:remove"
)

// Profile account operations.
const (
	ProfileAccountGetProfile                  = "profile:get"
	ProfileAccountGetRoles                    = "profile:get"
	ProfileAccountGetPermissions              = "profile:get"
	ProfileAccountUpdateProfile               = "profile:set"
	ProfileAccountUpdateMembershipPreferences = "profile:set"
)

// Profile security operations.
const (
	ProfileSecurityGetAuthMethods    = "profile_security:get"
	ProfileSecurityUpdateClassicAuth = "profile_security:set"
	ProfileSecurityChangePassword    = "profile_security:set"
)

// Client-origin administration within management client operations.
const (
	ManagementClientCreateOrigin          = "client_origin:add"
	ManagementClientReadOriginById        = "client_origin:get"
	ManagementClientReadOriginByValue     = "client_origin:get"
	ManagementClientReadOriginsByClientId = "client_origin:get"
	ManagementClientReadOriginsByFilter   = "client_origin:get"
	ManagementClientUpdateOriginById      = "client_origin:set"
	ManagementClientDeleteOriginById      = "client_origin:remove"
)
