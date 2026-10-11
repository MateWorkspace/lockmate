package presentationhttprequest

import (
	"encoding/json"
)

type ProfileUpdateProfileRequest struct {
	Name              *string          `json:"name" example:"Alex Doe"`
	Bio               *string          `json:"bio" example:"Example profile"`
	AvatarPath        *string          `json:"avatar_path" example:"avatars/alex.png"`
	AvatarPathSetNull bool             `json:"avatar_path_set_null" example:"true"`
	Preferences       *json.RawMessage `json:"preferences" swaggertype:"object"`
}

type ProfileUpdateMembershipPreferencesRequest struct {
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
}
