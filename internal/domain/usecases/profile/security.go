package domainusecasesprofile

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

// Security operates only on the caller's existing authentication methods.
// Classic credential operations return ClassicAuthNotFound for Google-only users.
type Security interface {
	GetAuthMethods(ctx context.Context, caller *domainmodelsentity.AccessClaims) (*AuthMethodsResult, error)
	// UpdateClassicAuth cannot set verification flags; changed contacts reset them.
	UpdateClassicAuth(ctx context.Context, caller *domainmodelsentity.AccessClaims, request UpdateClassicAuthRequest) error
	// ChangePassword verifies CurrentPassword first; mismatch returns PasswordMismatch.
	ChangePassword(ctx context.Context, caller *domainmodelsentity.AccessClaims, request ChangePasswordRequest) error
}

type AuthMethodsResult struct {
	ClassicAuth *domainusecasesshared.ClassicAuthView `json:"classic_auth,omitempty"`
	GoogleAuth  *domainmodelsentity.GoogleAuth        `json:"google_auth,omitempty"`
}

type UpdateClassicAuthRequest struct {
	Username *string

	Email        *string
	EmailSetNull bool
	Phone        *string
	PhoneSetNull bool
}

type ChangePasswordRequest struct {
	CurrentPassword string
	NewPassword     string
}
