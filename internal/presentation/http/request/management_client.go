package presentationhttprequest

import (
	"encoding/json"
)

type ManagementCreateClientRequest struct {
	Slug        string           `json:"slug" example:"web"`
	Name        string           `json:"name" example:"Alex Doe"`
	Description *string          `json:"description" example:"Example description"`
	IsActive    *bool            `json:"is_active" example:"true"`
	Preferences *json.RawMessage `json:"preferences" swaggertype:"object"`
}

type ManagementUpdateClientRequest struct {
	Name        *string          `json:"name" example:"Alex Doe"`
	Description *string          `json:"description" example:"Example description"`
	IsActive    *bool            `json:"is_active" example:"true"`
	Preferences *json.RawMessage `json:"preferences" swaggertype:"object"`
}

type ManagementCreateClientOriginRequest struct {
	Origin   string `json:"origin" example:"https://app.example.com"`
	IsActive *bool  `json:"is_active" example:"true"`
}

type ManagementUpdateClientOriginRequest struct {
	Origin   *string `json:"origin" example:"https://app.example.com"`
	IsActive *bool   `json:"is_active" example:"true"`
}
