package presentationhttpresponse

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type SpaceMemberRoleResponse struct {
	Id            int64 `json:"id" example:"1"`
	SpaceId       int64 `json:"space_id" example:"1"`
	SpaceMemberId int64 `json:"space_member_id" example:"1"`
	RoleId        int64 `json:"role_id" example:"1"`
	AuditCreateResponse
}

func SpaceMemberRole(value domainmodelsentity.SpaceMemberRole) SpaceMemberRoleResponse {
	return SpaceMemberRoleResponse{
		Id:                  value.Id,
		SpaceId:             value.SpaceId,
		SpaceMemberId:       value.SpaceMemberId,
		RoleId:              value.RoleId,
		AuditCreateResponse: AuditCreate(value.AuditCreate),
	}
}

func SpaceMemberRoleList(values []domainmodelsentity.SpaceMemberRole) []SpaceMemberRoleResponse {
	result := make([]SpaceMemberRoleResponse, 0, len(values))
	for _, value := range values {
		result = append(result, SpaceMemberRole(value))
	}
	return result
}
