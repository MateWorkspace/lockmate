package domainusecasesmanagement

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

// Client defines permission-protected administration; audit actors come from caller.
// Scoped operations require target-space or control-space authority.
type Client interface {
	Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request CreateClientRequest) (*ClientSecretResult, error)
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) (*domainusecasesshared.ClientView, error)
	ReadBySlug(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, value string) (*domainusecasesshared.ClientView, error)
	ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, request FilterClientRequest) (domainusecasesshared.Page[domainusecasesshared.ClientView], error)
	UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64, request UpdateClientRequest) error

	// RotateSecret returns the raw replacement once, after successful commit.
	RotateSecret(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, id int64) (*ClientSecretResult, error)
	DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId int64, id int64) error

	// IP allowlist administration always addresses a client within the target
	// space. Entries from another client return ClientIpAllowlistNotFound.
	CreateIpAllowlist(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, request CreateClientIpAllowlistRequest) (int64, error)
	ReadIpAllowlistById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64) (*domainmodelsentity.ClientIpAllowlist, error)
	ReadIpAllowlistByCidr(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, cidr string) (*domainmodelsentity.ClientIpAllowlist, error)
	ReadIpAllowlistsByClientId(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64) ([]domainmodelsentity.ClientIpAllowlist, error)
	ReadIpAllowlistsByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId int64, request FilterClientIpAllowlistRequest) (domainusecasesshared.Page[domainmodelsentity.ClientIpAllowlist], error)
	UpdateIpAllowlistById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64, request UpdateClientIpAllowlistRequest) error
	DeleteIpAllowlistById(ctx context.Context, caller *domainmodelsentity.AccessClaims, spaceId, clientId, id int64) error
}

type CreateClientRequest struct {
	AuthRedirectUri *string `json:"auth_redirect_uri" example:"https://app.example.com/auth/callback"`
	Slug            string
	Name            string
	Description     *string
	IsActive        *bool
	Preferences     *json.RawMessage
}

type FilterClientRequest struct {
	Page     int64
	Limit    int64
	Search   *string
	IsActive *bool
}

// UpdateClientRequest leaves omitted fields unchanged.
type UpdateClientRequest struct {
	AuthRedirectUri        *string `json:"auth_redirect_uri" example:"https://app.example.com/auth/callback"`
	SetAuthRedirectUriNull bool    `json:"auth_redirect_uri_set_null"`
	Name                   *string
	Description            *string
	IsActive               *bool
	Preferences            *json.RawMessage
}

// ClientSecretResult exposes a raw secret only at creation or rotation.
type ClientSecretResult struct {
	Client domainusecasesshared.ClientView `json:"client"`
	Secret string                          `json:"secret"`
}

type CreateClientIpAllowlistRequest struct {
	Cidr     string
	IsActive *bool
}

type UpdateClientIpAllowlistRequest struct {
	Cidr     *string
	IsActive *bool
}

type FilterClientIpAllowlistRequest struct {
	Page     int64
	Limit    int64
	Search   *string
	IsActive *bool
}
