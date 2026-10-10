package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type Permission interface {
	Create(ctx context.Context, spaceId int64, args CreatePermissionArgs) (id int64, err error)
	ReadById(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.Permission, error)
	ReadBySlug(ctx context.Context, spaceId int64, value string) (*domainmodelsentity.Permission, error)
	ReadByFilter(ctx context.Context, spaceId int64, filter PermissionFilter) (items []domainmodelsentity.Permission, total int64, err error)
	UpdateById(ctx context.Context, spaceId int64, id int64, args UpdatePermissionArgs) error
	DeleteById(ctx context.Context, spaceId int64, id int64, by *int64) error
}

type CreatePermissionArgs struct {
	Slug        string
	Name        string
	Description *string
	Preferences *json.RawMessage
	By          *int64
}

type PermissionFilter struct {
	Page   int64
	Limit  int64
	Search *string
}

type UpdatePermissionArgs struct {
	Name        *string
	Description *string
	Preferences *json.RawMessage
	By          *int64
}
