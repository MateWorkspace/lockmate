package presentationhttprequest

import (
	"encoding/json"
)

type ManagementCreateClassicAuthRequest struct {
	UserId          int64            `json:"user_id" example:"1"`
	Username        string           `json:"username" example:"alex"`
	Email           *string          `json:"email" example:"alex@example.com"`
	Phone           *string          `json:"phone" example:"+628123456789"`
	Password        string           `json:"password" example:"SecurePass123!"`
	IsEmailVerified *bool            `json:"is_email_verified" example:"true"`
	IsPhoneVerified *bool            `json:"is_phone_verified" example:"true"`
	Preferences     *json.RawMessage `json:"preferences" swaggertype:"object"`
}

type ManagementUpdateClassicAuthRequest struct {
	Username        *string          `json:"username" example:"alex"`
	Email           *string          `json:"email" example:"alex@example.com"`
	EmailSetNull    bool             `json:"email_set_null" example:"true"`
	Phone           *string          `json:"phone" example:"+628123456789"`
	PhoneSetNull    bool             `json:"phone_set_null" example:"true"`
	IsEmailVerified *bool            `json:"is_email_verified" example:"true"`
	IsPhoneVerified *bool            `json:"is_phone_verified" example:"true"`
	Preferences     *json.RawMessage `json:"preferences" swaggertype:"object"`
}

type ManagementResetUserPasswordRequest struct {
	Password string `json:"password" example:"SecurePass123!"`
}
