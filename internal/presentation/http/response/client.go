package presentationhttpresponse

import (
	"encoding/json"

	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

type ClientResponse struct {
	AuthRedirectUri   *string         `json:"auth_redirect_uri" example:"https://app.example.com/auth/callback"`
	Id                int64           `json:"id" example:"1"`
	SpaceId           int64           `json:"space_id" example:"1"`
	Slug              string          `json:"slug" example:"web"`
	Name              string          `json:"name" example:"Alex Doe"`
	Description       string          `json:"description" example:"Example description"`
	SecretKeyRedacted string          `json:"secret_key_redacted" example:"example"`
	IsActive          bool            `json:"is_active" example:"true"`
	Preferences       json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func Client(value domainusecasesshared.ClientView) ClientResponse {
	return ClientResponse{
		AuthRedirectUri:   value.AuthRedirectUri,
		Id:                value.Id,
		SpaceId:           value.SpaceId,
		Slug:              value.Slug,
		Name:              value.Name,
		Description:       value.Description,
		SecretKeyRedacted: value.SecretKeyRedacted,
		IsActive:          value.IsActive,
		Preferences:       NormalizeJSON(value.Preferences),
		AuditResponse:     Audit(value.AuditCreateUpdateDelete),
	}
}

func ClientList(values []domainusecasesshared.ClientView) []ClientResponse {
	result := make([]ClientResponse, 0, len(values))
	for _, value := range values {
		result = append(result, Client(value))
	}
	return result
}
