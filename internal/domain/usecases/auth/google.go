package domainusecasesauth

import (
	"context"

	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
)

// Google combines provider sign-in and first-time registration in one flow.
type Google interface {
	BeginSignIn(ctx context.Context, client *domainmodelsapp.ClientIdentity, request BeginGoogleSignInRequest) (*BeginGoogleSignInResult, error)
	// SignInCallback resolves client scope from the validated OAuth attempt.
	// Match by subject only; never link existing users by email coincidence.
	// Create missing user/identity/membership and synchronize Google contacts
	// atomically, returning tokens only after a successful commit.
	SignInCallback(ctx context.Context, request GoogleSignInCallbackRequest) (*SessionResult, error)
}

type BeginGoogleSignInRequest struct {
	// Binding comes from a trusted browser/session binding mechanism.
	Binding string
}

type BeginGoogleSignInResult struct {
	AuthorizationUrl string `json:"authorization_url"`
}

type GoogleSignInCallbackRequest struct {
	Code          string
	State         string
	Binding       string
	ProviderError string
}
