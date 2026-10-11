package domainusecasesshared

import domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"

// Joined read models are exposed through usecases without requiring callers to
// depend on repository contracts. These aliases preserve the existing API.
type SpaceMemberWithUser = domaincontractsrepository.SpaceMemberWithUser
type RolePermissionDetails = domaincontractsrepository.RolePermissionDetails
type RolePermissionWithRole = domaincontractsrepository.RolePermissionWithRole
type RolePermissionWithPermission = domaincontractsrepository.RolePermissionWithPermission
type SpaceMemberRoleDetails = domaincontractsrepository.SpaceMemberRoleDetails
type SpaceMemberRoleWithSpaceMember = domaincontractsrepository.SpaceMemberRoleWithSpaceMember
type SpaceMemberRoleWithRole = domaincontractsrepository.SpaceMemberRoleWithRole
