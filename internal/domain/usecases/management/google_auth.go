package domainusecasesmanagement

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

// GoogleAuth defines permission-protected administration; audit actors come from caller.
// Global operations require an authenticated control-space caller.
// Provider identities are monitoring-only; linking and mutations are not exposed.
type GoogleAuth interface {
	ReadById(ctx context.Context, caller *domainmodelsentity.AccessClaims, id int64) (*domainmodelsentity.GoogleAuth, error)
	ReadByUserId(ctx context.Context, caller *domainmodelsentity.AccessClaims, value int64) (*domainmodelsentity.GoogleAuth, error)
	ReadBySubject(ctx context.Context, caller *domainmodelsentity.AccessClaims, value string) (*domainmodelsentity.GoogleAuth, error)
	ReadByFilter(ctx context.Context, caller *domainmodelsentity.AccessClaims, request FilterGoogleAuthRequest) (domainusecasesshared.Page[domainmodelsentity.GoogleAuth], error)
}

type FilterGoogleAuthRequest struct {
	Page            int64
	Limit           int64
	Search          *string
	UserId          *int64
	IsEmailVerified *bool
}
