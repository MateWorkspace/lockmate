package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type Space interface {
	Create(ctx context.Context, args CreateSpaceArgs) (id int64, err error)
	ReadById(ctx context.Context, id int64) (*domainmodelsentity.Space, error)
	ReadBySlug(ctx context.Context, value string) (*domainmodelsentity.Space, error)
	ReadByFilter(ctx context.Context, filter SpaceFilter) (items []domainmodelsentity.Space, total int64, err error)
	UpdateById(ctx context.Context, id int64, args UpdateSpaceArgs) error
	DeleteById(ctx context.Context, id int64, by *int64) error
}

type CreateSpaceArgs struct {
	Slug        string
	Name        string
	Description *string
	IsActive    *bool
	Preferences *json.RawMessage
	By          *int64
}

type SpaceFilter struct {
	Page     int64
	Limit    int64
	Search   *string
	IsActive *bool
}

type UpdateSpaceArgs struct {
	Name        *string
	Description *string
	IsActive    *bool
	Preferences *json.RawMessage
	By          *int64
}
