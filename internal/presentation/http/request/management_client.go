package presentationhttprequest

import (
	"encoding/json"
)

type ManagementCreateClientRequest struct {
	AuthRedirectUri *string          `json:"auth_redirect_uri" example:"https://app.example.com/auth/callback"`
	Slug            string           `json:"slug" example:"web"`
	Name            string           `json:"name" example:"Alex Doe"`
	Description     *string          `json:"description" example:"Example description"`
	IsActive        *bool            `json:"is_active" example:"true"`
	Preferences     *json.RawMessage `json:"preferences" swaggertype:"object"`
}

type ManagementUpdateClientRequest struct {
	AuthRedirectUri        *string          `json:"auth_redirect_uri" example:"https://app.example.com/auth/callback"`
	SetAuthRedirectUriNull bool             `json:"auth_redirect_uri_set_null"`
	Name                   *string          `json:"name" example:"Alex Doe"`
	Description            *string          `json:"description" example:"Example description"`
	IsActive               *bool            `json:"is_active" example:"true"`
	Preferences            *json.RawMessage `json:"preferences" swaggertype:"object"`
}

type ManagementCreateClientOriginRequest struct {
	Origin   string `json:"origin" example:"https://app.example.com"`
	IsActive *bool  `json:"is_active" example:"true"`
}

type ManagementUpdateClientOriginRequest struct {
	Origin   *string `json:"origin" example:"https://app.example.com"`
	IsActive *bool   `json:"is_active" example:"true"`
}
