package presentationhttprequest

import (
	"encoding/json"
)

type ManagementCreateSpaceRequest struct {
	Slug        string           `json:"slug" example:"web"`
	Name        string           `json:"name" example:"Alex Doe"`
	Description *string          `json:"description" example:"Example description"`
	IsActive    *bool            `json:"is_active" example:"true"`
	Preferences *json.RawMessage `json:"preferences" swaggertype:"object"`
}

type ManagementUpdateSpaceRequest struct {
	Name        *string          `json:"name" example:"Alex Doe"`
	Description *string          `json:"description" example:"Example description"`
	IsActive    *bool            `json:"is_active" example:"true"`
	Preferences *json.RawMessage `json:"preferences" swaggertype:"object"`
}
