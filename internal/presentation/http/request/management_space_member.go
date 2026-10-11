package presentationhttprequest

import (
	"encoding/json"
)

type ManagementCreateSpaceMemberRequest struct {
	UserId      int64            `json:"user_id" example:"1"`
	IsActive    *bool            `json:"is_active" example:"true"`
	Preferences *json.RawMessage `json:"preferences" swaggertype:"object"`
}

type ManagementUpdateSpaceMemberRequest struct {
	IsActive    *bool            `json:"is_active" example:"true"`
	Preferences *json.RawMessage `json:"preferences" swaggertype:"object"`
}
