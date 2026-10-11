package presentationhttpresponse

import (
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type PermissionResponse struct {
	Id          int64           `json:"id" example:"1"`
	SpaceId     int64           `json:"space_id" example:"1"`
	Slug        string          `json:"slug" example:"web"`
	Name        string          `json:"name" example:"Alex Doe"`
	Description string          `json:"description" example:"Example description"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func Permission(value domainmodelsentity.Permission) PermissionResponse {
	return PermissionResponse{
		Id:            value.Id,
		SpaceId:       value.SpaceId,
		Slug:          value.Slug,
		Name:          value.Name,
		Description:   value.Description,
		Preferences:   NormalizeJSON(value.Preferences),
		AuditResponse: Audit(value.AuditCreateUpdateDelete),
	}
}

func PermissionList(values []domainmodelsentity.Permission) []PermissionResponse {
	result := make([]PermissionResponse, 0, len(values))
	for _, value := range values {
		result = append(result, Permission(value))
	}
	return result
}
