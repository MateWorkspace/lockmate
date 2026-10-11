package presentationhttpresponse

import (
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type UserResponse struct {
	Id          int64           `json:"id" example:"1"`
	Name        string          `json:"name" example:"Alex Doe"`
	Bio         string          `json:"bio" example:"Example profile"`
	AvatarPath  *string         `json:"avatar_path,omitempty" example:"avatars/alex.png"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func User(value domainmodelsentity.User) UserResponse {
	return UserResponse{
		Id:            value.Id,
		Name:          value.Name,
		Bio:           value.Bio,
		AvatarPath:    value.AvatarPath,
		Preferences:   NormalizeJSON(value.Preferences),
		AuditResponse: Audit(value.AuditCreateUpdateDelete),
	}
}

func UserList(values []domainmodelsentity.User) []UserResponse {
	result := make([]UserResponse, 0, len(values))
	for _, value := range values {
		result = append(result, User(value))
	}
	return result
}
