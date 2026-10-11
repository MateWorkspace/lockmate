package presentationhttpresponse

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type ClientOriginResponse struct {
	Id       int64  `json:"id" example:"1"`
	ClientId int64  `json:"client_id" example:"1"`
	Origin   string `json:"origin" example:"https://app.example.com"`
	IsActive bool   `json:"is_active" example:"true"`
	AuditResponse
}

func ClientOrigin(value domainmodelsentity.ClientOrigin) ClientOriginResponse {
	return ClientOriginResponse{
		Id:            value.Id,
		ClientId:      value.ClientId,
		Origin:        value.Origin,
		IsActive:      value.IsActive,
		AuditResponse: Audit(value.AuditCreateUpdateDelete),
	}
}

func ClientOriginList(values []domainmodelsentity.ClientOrigin) []ClientOriginResponse {
	result := make([]ClientOriginResponse, 0, len(values))
	for _, value := range values {
		result = append(result, ClientOrigin(value))
	}
	return result
}
