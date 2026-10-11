package presentationhttpresponse

import (
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type SpaceResponse struct {
	Id          int64           `json:"id" example:"1"`
	Slug        string          `json:"slug" example:"web"`
	Name        string          `json:"name" example:"Alex Doe"`
	Description string          `json:"description" example:"Example description"`
	IsActive    bool            `json:"is_active" example:"true"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func Space(value domainmodelsentity.Space) SpaceResponse {
	return SpaceResponse{
		Id:            value.Id,
		Slug:          value.Slug,
		Name:          value.Name,
		Description:   value.Description,
		IsActive:      value.IsActive,
		Preferences:   NormalizeJSON(value.Preferences),
		AuditResponse: Audit(value.AuditCreateUpdateDelete),
	}
}

func SpaceList(values []domainmodelsentity.Space) []SpaceResponse {
	result := make([]SpaceResponse, 0, len(values))
	for _, value := range values {
		result = append(result, Space(value))
	}
	return result
}
