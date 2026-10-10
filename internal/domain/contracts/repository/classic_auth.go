package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type ClassicAuth interface {
	Create(ctx context.Context, args CreateClassicAuthArgs) (id int64, err error)
	ReadById(ctx context.Context, id int64) (*domainmodelsentity.ClassicAuth, error)
	ReadByUserId(ctx context.Context, value int64) (*domainmodelsentity.ClassicAuth, error)
	ReadByUsername(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error)
	ReadByEmail(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error)
	ReadByPhone(ctx context.Context, value string) (*domainmodelsentity.ClassicAuth, error)
	ReadByFilter(ctx context.Context, filter ClassicAuthFilter) (items []ClassicAuthSummary, total int64, err error)
	UpdateById(ctx context.Context, id int64, args UpdateClassicAuthArgs) error
	DeleteById(ctx context.Context, id int64, by *int64) error
}

type CreateClassicAuthArgs struct {
	UserId          int64
	Username        string
	Email           *string
	Phone           *string
	PasswordHash    string
	IsEmailVerified *bool
	IsPhoneVerified *bool
	Preferences     *json.RawMessage
	By              *int64
}

type ClassicAuthFilter struct {
	Page            int64
	Limit           int64
	Search          *string
	UserId          *int64
	IsEmailVerified *bool
	IsPhoneVerified *bool
}

type UpdateClassicAuthArgs struct {
	Username        *string
	Email           *string
	EmailSetNull    bool
	Phone           *string
	PhoneSetNull    bool
	PasswordHash    *string
	IsEmailVerified *bool
	IsPhoneVerified *bool
	Preferences     *json.RawMessage
	By              *int64
}

// ClassicAuthSummary omits stored credential hashes from filtered results.
type ClassicAuthSummary struct {
	Id              int64           `db:"id" json:"id"`
	UserId          int64           `db:"user_id" json:"user_id"`
	Username        string          `db:"username" json:"username"`
	Email           *string         `db:"email" json:"email,omitempty"`
	Phone           *string         `db:"phone" json:"phone,omitempty"`
	IsEmailVerified bool            `db:"is_email_verified" json:"is_email_verified"`
	IsPhoneVerified bool            `db:"is_phone_verified" json:"is_phone_verified"`
	Preferences     json.RawMessage `db:"preferences" json:"preferences"`
	domainmodelsentity.AuditCreateUpdateDelete
}
