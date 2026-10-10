package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type User interface {
	Create(ctx context.Context, args CreateUserArgs) (id int64, err error)
	ReadById(ctx context.Context, id int64) (*domainmodelsentity.User, error)
	ReadByFilter(ctx context.Context, filter UserFilter) (items []domainmodelsentity.User, total int64, err error)
	UpdateById(ctx context.Context, id int64, args UpdateUserArgs) error
	DeleteById(ctx context.Context, id int64, by *int64) error
}

type CreateUserArgs struct {
	Name        string
	Bio         *string
	AvatarPath  *string
	Preferences *json.RawMessage
	By          *int64
}

type UserFilter struct {
	Page   int64
	Limit  int64
	Search *string
}

type UpdateUserArgs struct {
	Name              *string
	Bio               *string
	AvatarPath        *string
	AvatarPathSetNull bool
	Preferences       *json.RawMessage
	By                *int64
}
