package presentationhttpresponse

import (
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

type RolePermissionDetailsResponse struct {
	RolePermission RolePermissionResponse `json:"role_permission"`
	Role           RoleResponse           `json:"role"`
	Permission     PermissionResponse     `json:"permission"`
}

func RolePermissionDetails(value domainusecasesshared.RolePermissionDetails) RolePermissionDetailsResponse {
	return RolePermissionDetailsResponse{
		RolePermission: RolePermission(value.RolePermission),
		Role:           Role(value.Role),
		Permission:     Permission(value.Permission),
	}
}

func RolePermissionDetailsList(values []domainusecasesshared.RolePermissionDetails) []RolePermissionDetailsResponse {
	result := make([]RolePermissionDetailsResponse, 0, len(values))
	for _, value := range values {
		result = append(result, RolePermissionDetails(value))
	}
	return result
}

type RolePermissionWithRoleResponse struct {
	RolePermission RolePermissionResponse `json:"role_permission"`
	Role           RoleResponse           `json:"role"`
}

func RolePermissionWithRole(value domainusecasesshared.RolePermissionWithRole) RolePermissionWithRoleResponse {
	return RolePermissionWithRoleResponse{
		RolePermission: RolePermission(value.RolePermission),
		Role:           Role(value.Role),
	}
}

func RolePermissionWithRoleList(values []domainusecasesshared.RolePermissionWithRole) []RolePermissionWithRoleResponse {
	result := make([]RolePermissionWithRoleResponse, 0, len(values))
	for _, value := range values {
		result = append(result, RolePermissionWithRole(value))
	}
	return result
}

type RolePermissionWithPermissionResponse struct {
	RolePermission RolePermissionResponse `json:"role_permission"`
	Permission     PermissionResponse     `json:"permission"`
}

func RolePermissionWithPermission(value domainusecasesshared.RolePermissionWithPermission) RolePermissionWithPermissionResponse {
	return RolePermissionWithPermissionResponse{
		RolePermission: RolePermission(value.RolePermission),
		Permission:     Permission(value.Permission),
	}
}

func RolePermissionWithPermissionList(values []domainusecasesshared.RolePermissionWithPermission) []RolePermissionWithPermissionResponse {
	result := make([]RolePermissionWithPermissionResponse, 0, len(values))
	for _, value := range values {
		result = append(result, RolePermissionWithPermission(value))
	}
	return result
}
