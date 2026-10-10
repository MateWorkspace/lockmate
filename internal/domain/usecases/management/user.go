package domainusecasesmanagement

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

// User defines permission-protected administration; audit actors come from caller.
// Global operations require an authenticated control-space caller.
// Create atomically provisions the user and classic credentials, without enrollment.
type User interface {
	Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, request CreateUserRequest) (id int64, err error)
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (*domainmodelsentity.User, error)
	ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request FilterUserRequest) (domainusecasesshared.Page[domainmodelsentity.User], error)
	UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64, request UpdateUserRequest) error
	DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) error
}

type CreateUserRequest struct {
	Name        string
	Bio         *string
	AvatarPath  *string
	Preferences *json.RawMessage

	Username        string
	Password        string
	Email           *string
	Phone           *string
	IsEmailVerified *bool
	IsPhoneVerified *bool
}

type FilterUserRequest struct {
	Page   int64
	Limit  int64
	Search *string
}

// UpdateUserRequest leaves omitted fields unchanged.
type UpdateUserRequest struct {
	Name              *string
	Bio               *string
	AvatarPath        *string
	AvatarPathSetNull bool
	Preferences       *json.RawMessage
}
