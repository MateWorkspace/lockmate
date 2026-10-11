package presentationhttpresponse

import (
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type GoogleAuthResponse struct {
	Id              int64           `json:"id" example:"1"`
	UserId          int64           `json:"user_id" example:"1"`
	Subject         string          `json:"subject" example:"123456789012345678901"`
	Email           string          `json:"email" example:"alex@example.com"`
	IsEmailVerified bool            `json:"is_email_verified" example:"true"`
	Preferences     json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func GoogleAuth(value domainmodelsentity.GoogleAuth) GoogleAuthResponse {
	return GoogleAuthResponse{
		Id:              value.Id,
		UserId:          value.UserId,
		Subject:         value.Subject,
		Email:           value.Email,
		IsEmailVerified: value.IsEmailVerified,
		Preferences:     NormalizeJSON(value.Preferences),
		AuditResponse:   Audit(value.AuditCreateUpdateDelete),
	}
}

func GoogleAuthList(values []domainmodelsentity.GoogleAuth) []GoogleAuthResponse {
	result := make([]GoogleAuthResponse, 0, len(values))
	for _, value := range values {
		result = append(result, GoogleAuth(value))
	}
	return result
}
