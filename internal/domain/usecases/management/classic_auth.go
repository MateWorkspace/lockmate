package domainusecasesmanagement

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

// ClassicAuth defines permission-protected administration; audit actors come from caller.
// Global operations require an authenticated control-space caller.
type ClassicAuth interface {
	Create(ctx context.Context, caller *domainmodelsentity.AccessClaims, request CreateClassicAuthRequest) (id int64, err error)
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (*domainusecasesshared.ClassicAuthView, error)
	ReadByUserId(ctx context.Context, caller *domainmodelsentity.AccessClaims, value int64) (*domainusecasesshared.ClassicAuthView, error)
	ReadByUsername(ctx context.Context, caller *domainmodelsentity.AccessClaims, value string) (*domainusecasesshared.ClassicAuthView, error)
	ReadByEmail(ctx context.Context, caller *domainmodelsentity.AccessClaims, value string) (*domainusecasesshared.ClassicAuthView, error)
	ReadByPhone(ctx context.Context, caller *domainmodelsentity.AccessClaims, value string) (*domainusecasesshared.ClassicAuthView, error)
	ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request FilterClassicAuthRequest) (domainusecasesshared.Page[domainusecasesshared.ClassicAuthView], error)
	UpdateById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64, request UpdateClassicAuthRequest) error

	// ResetPassword replaces existing credentials without requiring the old password.
	ResetPassword(ctx context.Context, caller *domainmodelsentity.AccessClaims, userId int64, request ResetUserPasswordRequest) error

	// DeleteById requires a remaining live Google identity; otherwise return BadState.
	DeleteById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) error
}

type CreateClassicAuthRequest struct {
	UserId          int64
	Username        string
	Email           *string
	Phone           *string
	Password        string
	IsEmailVerified *bool
	IsPhoneVerified *bool
	Preferences     *json.RawMessage
}

type FilterClassicAuthRequest struct {
	Page            int64
	Limit           int64
	Search          *string
	UserId          *int64
	IsEmailVerified *bool
	IsPhoneVerified *bool
}

// UpdateClassicAuthRequest leaves omitted fields unchanged.
type UpdateClassicAuthRequest struct {
	Username        *string
	Email           *string
	EmailSetNull    bool
	Phone           *string
	PhoneSetNull    bool
	IsEmailVerified *bool
	IsPhoneVerified *bool
	Preferences     *json.RawMessage
}

type ResetUserPasswordRequest struct {
	Password string
}
