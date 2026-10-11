package presentationhttpresponse

import (
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type RoleResponse struct {
	Id          int64           `json:"id" example:"1"`
	SpaceId     int64           `json:"space_id" example:"1"`
	Slug        string          `json:"slug" example:"web"`
	Name        string          `json:"name" example:"Alex Doe"`
	Description string          `json:"description" example:"Example description"`
	IsDefault   bool            `json:"is_default" example:"true"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func Role(value domainmodelsentity.Role) RoleResponse {
	return RoleResponse{
		Id:            value.Id,
		SpaceId:       value.SpaceId,
		Slug:          value.Slug,
		Name:          value.Name,
		Description:   value.Description,
		IsDefault:     value.IsDefault,
		Preferences:   NormalizeJSON(value.Preferences),
		AuditResponse: Audit(value.AuditCreateUpdateDelete),
	}
}

func RoleList(values []domainmodelsentity.Role) []RoleResponse {
	result := make([]RoleResponse, 0, len(values))
	for _, value := range values {
		result = append(result, Role(value))
	}
	return result
}
