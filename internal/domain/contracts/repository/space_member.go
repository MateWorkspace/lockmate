package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type SpaceMember interface {
	Create(ctx context.Context, spaceId int64, args CreateSpaceMemberArgs) (id int64, err error)
	ReadById(ctx context.Context, spaceId int64, id int64) (*domainmodelsentity.SpaceMember, error)
	ReadByUserId(ctx context.Context, spaceId int64, value int64) (*domainmodelsentity.SpaceMember, error)
	ReadByFilter(ctx context.Context, spaceId int64, filter SpaceMemberFilter) (items []SpaceMemberWithUser, total int64, err error)
	UpdateById(ctx context.Context, spaceId int64, id int64, args UpdateSpaceMemberArgs) error
	DeleteById(ctx context.Context, spaceId int64, id int64, by *int64) error
}

type CreateSpaceMemberArgs struct {
	UserId      int64
	IsActive    *bool
	Preferences *json.RawMessage
	By          *int64
}

type SpaceMemberFilter struct {
	Page     int64
	Limit    int64
	Search   *string
	UserId   *int64
	RoleId   *int64
	IsActive *bool
}

type UpdateSpaceMemberArgs struct {
	IsActive    *bool
	Preferences *json.RawMessage
	By          *int64
}

type SpaceMemberWithUser struct {
	SpaceMember domainmodelsentity.SpaceMember
	User        domainmodelsentity.User
}
