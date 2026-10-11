package presentationhttprequest

import (
	"encoding/json"
)

type ManagementCreateUserRequest struct {
	Name            string           `json:"name" example:"Alex Doe"`
	Bio             *string          `json:"bio" example:"Example profile"`
	AvatarPath      *string          `json:"avatar_path" example:"avatars/alex.png"`
	Preferences     *json.RawMessage `json:"preferences" swaggertype:"object"`
	Username        string           `json:"username" example:"alex"`
	Password        string           `json:"password" example:"SecurePass123!"`
	Email           *string          `json:"email" example:"alex@example.com"`
	Phone           *string          `json:"phone" example:"+628123456789"`
	IsEmailVerified *bool            `json:"is_email_verified" example:"true"`
	IsPhoneVerified *bool            `json:"is_phone_verified" example:"true"`
}

type ManagementUpdateUserRequest struct {
	Name              *string          `json:"name" example:"Alex Doe"`
	Bio               *string          `json:"bio" example:"Example profile"`
	AvatarPath        *string          `json:"avatar_path" example:"avatars/alex.png"`
	AvatarPathSetNull bool             `json:"avatar_path_set_null" example:"true"`
	Preferences       *json.RawMessage `json:"preferences" swaggertype:"object"`
}
