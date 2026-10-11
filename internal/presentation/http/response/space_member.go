package presentationhttpresponse

import (
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type SpaceMemberResponse struct {
	Id          int64           `json:"id" example:"1"`
	SpaceId     int64           `json:"space_id" example:"1"`
	UserId      int64           `json:"user_id" example:"1"`
	IsActive    bool            `json:"is_active" example:"true"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func SpaceMember(value domainmodelsentity.SpaceMember) SpaceMemberResponse {
	return SpaceMemberResponse{
		Id:            value.Id,
		SpaceId:       value.SpaceId,
		UserId:        value.UserId,
		IsActive:      value.IsActive,
		Preferences:   NormalizeJSON(value.Preferences),
		AuditResponse: Audit(value.AuditCreateUpdateDelete),
	}
}

func SpaceMemberList(values []domainmodelsentity.SpaceMember) []SpaceMemberResponse {
	result := make([]SpaceMemberResponse, 0, len(values))
	for _, value := range values {
		result = append(result, SpaceMember(value))
	}
	return result
}
