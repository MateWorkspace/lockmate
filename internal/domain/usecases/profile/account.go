package domainusecasesprofile

import (
	"context"
	"encoding/json"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// Account derives user, space, membership, and audit IDs from authenticated claims.
// Global profile edits affect the user across all spaces.
type Account interface {
	GetProfile(ctx context.Context, caller *domainmodelsentity.AccessClaims) (*ProfileResult, error)
	GetRoles(ctx context.Context, caller *domainmodelsentity.AccessClaims) ([]domainmodelsentity.Role, error)
	GetPermissions(ctx context.Context, caller *domainmodelsentity.AccessClaims) ([]domainmodelsentity.Permission, error)
	UpdateProfile(ctx context.Context, caller *domainmodelsentity.AccessClaims, request UpdateProfileRequest) error
	UpdateMembershipPreferences(ctx context.Context, caller *domainmodelsentity.AccessClaims, request UpdateMembershipPreferencesRequest) error
}

type ProfileResult struct {
	User        domainmodelsentity.User        `json:"user"`
	Space       domainmodelsentity.Space       `json:"space"`
	SpaceMember domainmodelsentity.SpaceMember `json:"space_member"`
}

type UpdateProfileRequest struct {
	Name *string
	Bio  *string

	// AvatarPath changes metadata only; there is no object-storage operation.
	AvatarPath        *string
	AvatarPathSetNull bool
	Preferences       *json.RawMessage
}

type UpdateMembershipPreferencesRequest struct {
	Preferences json.RawMessage
}
