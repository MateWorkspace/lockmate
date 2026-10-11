package presentationhttprequest

import (
	"encoding/json"
)

type ManagementCreatePermissionRequest struct {
	Slug        string           `json:"slug" example:"web"`
	Name        string           `json:"name" example:"Alex Doe"`
	Description *string          `json:"description" example:"Example description"`
	Preferences *json.RawMessage `json:"preferences" swaggertype:"object"`
}

type ManagementUpdatePermissionRequest struct {
	Name        *string          `json:"name" example:"Alex Doe"`
	Description *string          `json:"description" example:"Example description"`
	Preferences *json.RawMessage `json:"preferences" swaggertype:"object"`
}
