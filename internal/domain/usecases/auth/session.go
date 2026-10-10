package domainusecasesauth

import (
	"context"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// Session derives space scope from a trusted client authentication result.
// Every operation rechecks that client and its active parent space.
type Session interface {
	// SignIn verifies classic credentials before enrolling a missing membership.
	// Suspended memberships are never reactivated by authentication.
	SignIn(ctx context.Context, client *domainmodelsentity.ClientIdentity, request SignInRequest) (*SessionResult, error)
	// Register creates user, credentials, membership, and default role atomically.
	// Return generated tokens only after the transaction commits successfully.
	Register(ctx context.Context, client *domainmodelsentity.ClientIdentity, request RegisterRequest) (*SessionResult, error)
	// Refresh requires an existing active membership and reloads current grants.
	// It issues a new pair without revoking the previously issued refresh token.
	Refresh(ctx context.Context, client *domainmodelsentity.ClientIdentity, request RefreshRequest) (*SessionResult, error)
	// AuthenticateAccessToken validates the token, rechecks live identity and
	// membership ownership, and returns current grants. It never enrolls users.
	AuthenticateAccessToken(ctx context.Context, client *domainmodelsentity.ClientIdentity, accessToken string) (*domainmodelsentity.AccessClaims, error)
}

type SignInRequest struct {
	Username string
	Password string
}

type RegisterRequest struct {
	Name     string
	Username string
	Password string

	Bio   *string
	Email *string
	Phone *string
}

type RefreshRequest struct {
	RefreshToken string
}

// SessionResult contains application tokens, never credential hashes or provider tokens.
type SessionResult struct {
	User        domainmodelsentity.User         `json:"user"`
	Space       domainmodelsentity.Space        `json:"space"`
	SpaceMember domainmodelsentity.SpaceMember  `json:"space_member"`
	Roles       []domainmodelsentity.Role       `json:"roles"`
	Permissions []domainmodelsentity.Permission `json:"permissions"`

	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IsNewUser    bool   `json:"is_new_user"`
}
