package domaincontractsutility

import "context"

// GoogleOAuth owns provider verification and the transient authorization attempt.
// Implementations use a configured callback URL, expiring single-use state,
// nonce, PKCE, and a browser/session binding supplied by trusted presentation code.
// No provider tokens or unverified identity fields may escape this boundary.
type GoogleOAuth interface {
	Begin(ctx context.Context, args BeginGoogleOAuthArgs) (authorizationUrl string, err error)
	// Complete consumes the attempt even when Google denies authorization.
	// Validate state and browser binding before exchanging or accepting a code.
	Complete(ctx context.Context, args CompleteGoogleOAuthArgs) (*CompleteGoogleOAuthResult, error)
}

type BeginGoogleOAuthArgs struct {
	ClientId int64
	SpaceId  int64
	// Binding is an opaque browser/session secret, not the returned state value.
	Binding string
}

type CompleteGoogleOAuthArgs struct {
	Code          string
	State         string
	Binding       string
	ProviderError string
}

type CompleteGoogleOAuthResult struct {
	// Client and space are recovered from the validated original attempt.
	ClientId int64
	SpaceId  int64
	Identity GoogleOAuthIdentity
}

type GoogleOAuthIdentity struct {
	// Subject is the verified, stable Google sub claim; email is not an identifier.
	Subject         string
	Email           string
	IsEmailVerified bool
	Name            *string
}
