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

type ManagementCreateClientIpAllowlistRequest struct {
	Cidr     string `json:"cidr" example:"203.0.113.0/24"`
	IsActive *bool  `json:"is_active" example:"true"`
}

type ManagementUpdateClientIpAllowlistRequest struct {
	Cidr     *string `json:"cidr" example:"203.0.113.0/24"`
	IsActive *bool   `json:"is_active" example:"true"`
}
