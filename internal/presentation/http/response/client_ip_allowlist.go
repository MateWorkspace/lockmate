package presentationhttpresponse

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type ClientIpAllowlistResponse struct {
	Id       int64  `json:"id" example:"1"`
	ClientId int64  `json:"client_id" example:"1"`
	Cidr     string `json:"cidr" example:"203.0.113.0/24"`
	IsActive bool   `json:"is_active" example:"true"`
	AuditResponse
}

func ClientIpAllowlist(value domainmodelsentity.ClientIpAllowlist) ClientIpAllowlistResponse {
	return ClientIpAllowlistResponse{
		Id:            value.Id,
		ClientId:      value.ClientId,
		Cidr:          value.Cidr,
		IsActive:      value.IsActive,
		AuditResponse: Audit(value.AuditCreateUpdateDelete),
	}
}

func ClientIpAllowlistList(values []domainmodelsentity.ClientIpAllowlist) []ClientIpAllowlistResponse {
	result := make([]ClientIpAllowlistResponse, 0, len(values))
	for _, value := range values {
		result = append(result, ClientIpAllowlist(value))
	}
	return result
}
