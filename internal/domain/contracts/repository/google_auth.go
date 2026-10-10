package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

type GoogleAuth interface {
	Create(ctx context.Context, args CreateGoogleAuthArgs) (id int64, err error)
	ReadById(ctx context.Context, id int64) (*domainmodelsentity.GoogleAuth, error)
	ReadByUserId(ctx context.Context, value int64) (*domainmodelsentity.GoogleAuth, error)
	ReadBySubject(ctx context.Context, value string) (*domainmodelsentity.GoogleAuth, error)
	ReadByFilter(ctx context.Context, filter GoogleAuthFilter) (items []domainmodelsentity.GoogleAuth, total int64, err error)
	UpdateById(ctx context.Context, id int64, args UpdateGoogleAuthArgs) error
	DeleteById(ctx context.Context, id int64, by *int64) error
}

type CreateGoogleAuthArgs struct {
	UserId          int64
	Subject         string
	Email           string
	IsEmailVerified *bool
	Preferences     *json.RawMessage
	By              *int64
}

type GoogleAuthFilter struct {
	Page            int64
	Limit           int64
	Search          *string
	UserId          *int64
	IsEmailVerified *bool
}

type UpdateGoogleAuthArgs struct {
	Email           *string
	IsEmailVerified *bool
	Preferences     *json.RawMessage
	By              *int64
}
