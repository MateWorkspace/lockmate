package presentationhttpresponse

import (
	"encoding/json"

	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

type ClassicAuthResponse struct {
	Id              int64           `json:"id" example:"1"`
	UserId          int64           `json:"user_id" example:"1"`
	Username        string          `json:"username" example:"alex"`
	Email           *string         `json:"email,omitempty" example:"alex@example.com"`
	Phone           *string         `json:"phone,omitempty" example:"+628123456789"`
	IsEmailVerified bool            `json:"is_email_verified" example:"true"`
	IsPhoneVerified bool            `json:"is_phone_verified" example:"true"`
	Preferences     json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func ClassicAuth(value domainusecasesshared.ClassicAuthView) ClassicAuthResponse {
	return ClassicAuthResponse{
		Id:              value.Id,
		UserId:          value.UserId,
		Username:        value.Username,
		Email:           value.Email,
		Phone:           value.Phone,
		IsEmailVerified: value.IsEmailVerified,
		IsPhoneVerified: value.IsPhoneVerified,
		Preferences:     NormalizeJSON(value.Preferences),
		AuditResponse:   Audit(value.AuditCreateUpdateDelete),
	}
}

func ClassicAuthList(values []domainusecasesshared.ClassicAuthView) []ClassicAuthResponse {
	result := make([]ClassicAuthResponse, 0, len(values))
	for _, value := range values {
		result = append(result, ClassicAuth(value))
	}
	return result
}
