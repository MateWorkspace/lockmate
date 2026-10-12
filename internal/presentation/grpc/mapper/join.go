package presentationgrpcmapper

import (
	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

func SpaceMemberWithUser(value domainusecasesshared.SpaceMemberWithUser) *lockmatev1.SpaceMemberWithUser {
	return &lockmatev1.SpaceMemberWithUser{
		SpaceMember: SpaceMember(value.SpaceMember),
		User:        User(value.User),
	}
}

func SpaceMemberWithUsers(values []domainusecasesshared.SpaceMemberWithUser) []*lockmatev1.SpaceMemberWithUser {
	result := make([]*lockmatev1.SpaceMemberWithUser, 0, len(values))
	for _, value := range values {
		result = append(result, SpaceMemberWithUser(value))
	}
	return result
}

func RolePermissionDetails(value domainusecasesshared.RolePermissionDetails) *lockmatev1.RolePermissionDetails {
	return &lockmatev1.RolePermissionDetails{
		RolePermission: RolePermission(value.RolePermission),
		Role:           Role(value.Role),
		Permission:     Permission(value.Permission),
	}
}

func RolePermissionWithRoles(values []domainusecasesshared.RolePermissionWithRole) []*lockmatev1.RolePermissionWithRole {
	result := make([]*lockmatev1.RolePermissionWithRole, 0, len(values))
	for _, value := range values {
		result = append(result, &lockmatev1.RolePermissionWithRole{
			RolePermission: RolePermission(value.RolePermission),
			Role:           Role(value.Role),
		})
	}
	return result
}

func RolePermissionWithPermissions(values []domainusecasesshared.RolePermissionWithPermission) []*lockmatev1.RolePermissionWithPermission {
	result := make([]*lockmatev1.RolePermissionWithPermission, 0, len(values))
	for _, value := range values {
		result = append(result, &lockmatev1.RolePermissionWithPermission{
			RolePermission: RolePermission(value.RolePermission),
			Permission:     Permission(value.Permission),
		})
	}
	return result
}

func SpaceMemberRoleDetails(value domainusecasesshared.SpaceMemberRoleDetails) *lockmatev1.SpaceMemberRoleDetails {
	return &lockmatev1.SpaceMemberRoleDetails{
		SpaceMemberRole: SpaceMemberRole(value.SpaceMemberRole),
		SpaceMember:     SpaceMember(value.SpaceMember),
		Role:            Role(value.Role),
	}
}

func SpaceMemberRoleWithSpaceMembers(values []domainusecasesshared.SpaceMemberRoleWithSpaceMember) []*lockmatev1.SpaceMemberRoleWithSpaceMember {
	result := make([]*lockmatev1.SpaceMemberRoleWithSpaceMember, 0, len(values))
	for _, value := range values {
		result = append(result, &lockmatev1.SpaceMemberRoleWithSpaceMember{
			SpaceMemberRole: SpaceMemberRole(value.SpaceMemberRole),
			SpaceMember:     SpaceMember(value.SpaceMember),
		})
	}
	return result
}

func SpaceMemberRoleWithRoles(values []domainusecasesshared.SpaceMemberRoleWithRole) []*lockmatev1.SpaceMemberRoleWithRole {
	result := make([]*lockmatev1.SpaceMemberRoleWithRole, 0, len(values))
	for _, value := range values {
		result = append(result, &lockmatev1.SpaceMemberRoleWithRole{
			SpaceMemberRole: SpaceMemberRole(value.SpaceMemberRole),
			Role:            Role(value.Role),
		})
	}
	return result
}
