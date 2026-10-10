package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type Role interface {
	Create(ctx context.Context, spaceId int64, args CreateRoleArgs) (id int64, err error)
	ReadById(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.Role, error)
	ReadBySlug(ctx context.Context, spaceId int64, value string) (*domainmodelsentity.Role, error)
	ReadByFilter(ctx context.Context, spaceId int64, filter RoleFilter) (items []domainmodelsentity.Role, total int64, err error)
	ReadDefault(ctx context.Context, spaceId int64) (*domainmodelsentity.Role, error)
	UpdateById(ctx context.Context, spaceId int64, id int64, args UpdateRoleArgs) error
	DeleteById(ctx context.Context, spaceId int64, id int64, by *int64) error
}

type CreateRoleArgs struct {
	Slug        string
	Name        string
	Description *string
	IsDefault   *bool
	Preferences *json.RawMessage
	By          *int64
}

type RoleFilter struct {
	Page      int64
	Limit     int64
	Search    *string
	IsDefault *bool
}

type UpdateRoleArgs struct {
	Name        *string
	Description *string
	IsDefault   *bool
	Preferences *json.RawMessage
	By          *int64
}
