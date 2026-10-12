package domainusecasesauth

import (
	"context"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

// Google combines verified provider sign-in and first-time registration.
// Callback creates a handoff; JWTs are issued only by Handoff.Exchange.
type Google interface {
	Start(ctx context.Context, client *domainmodelsentity.ClientIdentity, request StartGoogleSignInRequest) (*StartGoogleSignInResult, error)
	Launch(ctx context.Context, ticket string) (*LaunchGoogleSignInResult, error)
	SignInCallback(ctx context.Context, request GoogleSignInCallbackRequest) (*GoogleSignInCallbackResult, error)
}
type StartGoogleSignInRequest struct{ State string }
type StartGoogleSignInResult struct {
	LaunchUrl string `json:"launch_url"`
}

// Binding and state are internal browser-cookie data, never JSON response fields.
type LaunchGoogleSignInResult struct {
	AuthorizationUrl string
	Binding          string
	State            string
	CookieDuration   int
}
type GoogleSignInCallbackRequest struct {
	Code          string
	State         string
	Binding       string
	ProviderError string
}
type GoogleSignInCallbackResult struct{ RedirectUrl string }
