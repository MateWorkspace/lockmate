package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type Client interface {
	Create(ctx context.Context, spaceId int64, args CreateClientArgs) (id int64, err error)
	ReadById(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.Client, error)
	ReadBySlug(ctx context.Context, value string) (*domainmodelsentity.Client, error)
	ReadByFilter(ctx context.Context, spaceId int64, filter ClientFilter) (items []ClientSummary, total int64, err error)
	UpdateById(ctx context.Context, spaceId int64, id int64, args UpdateClientArgs) error
	DeleteById(ctx context.Context, spaceId int64, id int64, by *int64) error
}

type CreateClientArgs struct {
	AuthRedirectUri   *string
	Slug              string
	Name              string
	Description       *string
	SecretKeyHash     string
	SecretKeyRedacted string
	IsActive          *bool
	Preferences       *json.RawMessage
	By                *int64
}

type ClientFilter struct {
	Page     int64
	Limit    int64
	Search   *string
	IsActive *bool
}

type UpdateClientArgs struct {
	AuthRedirectUri        *string
	SetAuthRedirectUriNull bool
	Name                   *string
	Description            *string
	SecretKeyHash          *string
	SecretKeyRedacted      *string
	IsActive               *bool
	Preferences            *json.RawMessage
	By                     *int64
}

// ClientSummary omits stored credential hashes from filtered results.
type ClientSummary struct {
	Id                int64           `db:"id" json:"id"`
	SpaceId           int64           `db:"space_id" json:"space_id"`
	Slug              string          `db:"slug" json:"slug"`
	Name              string          `db:"name" json:"name"`
	Description       string          `db:"description" json:"description"`
	AuthRedirectUri   *string         `db:"auth_redirect_uri" json:"auth_redirect_uri"`
	SecretKeyRedacted string          `db:"secret_key_redacted" json:"secret_key_redacted"`
	IsActive          bool            `db:"is_active" json:"is_active"`
	Preferences       json.RawMessage `db:"preferences" json:"preferences"`
	domainmodelsentity.AuditCreateUpdateDelete
}
