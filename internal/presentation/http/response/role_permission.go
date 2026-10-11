package presentationhttpresponse

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type RolePermissionResponse struct {
	Id           int64 `json:"id" example:"1"`
	SpaceId      int64 `json:"space_id" example:"1"`
	RoleId       int64 `json:"role_id" example:"1"`
	PermissionId int64 `json:"permission_id" example:"1"`
	AuditCreateResponse
}

func RolePermission(value domainmodelsentity.RolePermission) RolePermissionResponse {
	return RolePermissionResponse{
		Id:                  value.Id,
		SpaceId:             value.SpaceId,
		RoleId:              value.RoleId,
		PermissionId:        value.PermissionId,
		AuditCreateResponse: AuditCreate(value.AuditCreate),
	}
}

func RolePermissionList(values []domainmodelsentity.RolePermission) []RolePermissionResponse {
	result := make([]RolePermissionResponse, 0, len(values))
	for _, value := range values {
		result = append(result, RolePermission(value))
	}
	return result
}
