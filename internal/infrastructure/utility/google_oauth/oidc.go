package infrastructureutilitygoogleoauth

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Config contains provider settings. Endpoint overrides support isolated provider tests;
// composition uses the Google defaults and never accepts endpoints from requests.
type Config struct {
	Enabled         bool
	ClientId        string
	ClientSecret    string
	CallbackUri     string
	AttemptDuration time.Duration
	Issuer          string
	KeysUrl         string
	Endpoint        oauth2.Endpoint
}

type oidcImpl struct {
	attempts domaincontractsrepository.GoogleOAuthAttempt
	flow     domaincontractsutility.AuthFlow
	client   *http.Client
	config   Config
	oauth    oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewOIDCImpl(attempts domaincontractsrepository.GoogleOAuthAttempt, flow domaincontractsutility.AuthFlow, client *http.Client, config Config) domaincontractsutility.GoogleOAuth {
	if config.Issuer == "" {
		config.Issuer = "https://accounts.google.com"
	}

	if config.KeysUrl == "" {
		config.KeysUrl = "https://www.googleapis.com/oauth2/v3/certs"
	}

	if config.Endpoint.AuthURL == "" {
		config.Endpoint = google.Endpoint
	}

	providerCtx := oidc.ClientContext(context.Background(), client)
	return &oidcImpl{
		attempts: attempts,
		flow:     flow,
		client:   client,
		config:   config,
		oauth: oauth2.Config{ClientID: config.ClientId,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.CallbackUri,
			Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
			Endpoint:     config.Endpoint},
		verifier: oidc.NewVerifier(config.Issuer, oidc.NewRemoteKeySet(providerCtx, config.KeysUrl), &oidc.Config{ClientID: config.ClientId}),
	}
}

func (o *oidcImpl) Begin(ctx context.Context, args domaincontractsutility.BeginGoogleOAuthArgs) (*domaincontractsutility.BeginGoogleOAuthResult, error) {
	if !o.config.Enabled {
		return nil, domainmodelserror.NewError("Google authentication is disabled", domainmodelserror.Unimplemented, nil)
	}

	if args.ClientId <= 0 || args.SpaceId <= 0 || args.Binding == "" || args.RedirectUri == "" || args.ClientState == "" {
		return nil, domainmodelserror.NewError("invalid Google authorization scope", domainmodelserror.BadArgs, nil)
	}

	state, err := o.flow.Generate()
	if err != nil {
		return nil, err
	}

	nonce, err := o.flow.Generate()
	if err != nil {
		return nil, err
	}

	verifier, err := o.flow.Generate()
	if err != nil {
		return nil, err
	}

	if err := o.attempts.Create(ctx, domainmodelsentity.GoogleOAuthAttempt{
		StateHash:   o.flow.Hash(state),
		BindingHash: o.flow.Hash(args.Binding),
		NonceHash:   o.flow.Hash(nonce),
		Verifier:    verifier,
		ClientId:    args.ClientId,
		SpaceId:     args.SpaceId,
		RedirectUri: args.RedirectUri,
		ClientState: args.ClientState,
		ExpiresAt:   time.Now().Add(o.config.AttemptDuration),
	}); err != nil {
		return nil, err
	}

	return &domaincontractsutility.BeginGoogleOAuthResult{
		State:            state,
		AuthorizationUrl: o.oauth.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oidc.Nonce(nonce)),
	}, nil
}

func (o *oidcImpl) Complete(ctx context.Context, args domaincontractsutility.CompleteGoogleOAuthArgs) (*domaincontractsutility.CompleteGoogleOAuthResult, error) {
	if !o.config.Enabled {
		return nil, domainmodelserror.NewError("Google authentication is disabled", domainmodelserror.Unimplemented, nil)
	}

	if args.State == "" || args.Binding == "" || (args.Code == "") == (args.ProviderError == "") {
		return nil, domainmodelserror.NewError("invalid Google callback", domainmodelserror.GoogleOAuthStateInvalid, nil)
	}

	attempt, err := o.attempts.Consume(ctx, o.flow.Hash(args.State), o.flow.Hash(args.Binding))
	if err != nil {
		return nil, err
	}

	if attempt == nil {
		return nil, domainmodelserror.NewError("Google authorization attempt is missing", domainmodelserror.Failure, nil)
	}

	result := &domaincontractsutility.CompleteGoogleOAuthResult{
		ClientId:    attempt.ClientId,
		SpaceId:     attempt.SpaceId,
		RedirectUri: attempt.RedirectUri,
		ClientState: attempt.ClientState,
	}
	if args.ProviderError != "" {
		result.Denied = true
		return result, nil
	}

	exchangeCtx := context.WithValue(ctx, oauth2.HTTPClient, o.client)
	token, err := o.oauth.Exchange(exchangeCtx, args.Code, oauth2.VerifierOption(attempt.Verifier))
	if err != nil {
		return nil, providerError("Google authorization exchange failed", err)
	}

	raw, ok := token.Extra("id_token").(string)
	if !ok || raw == "" {
		return nil, domainmodelserror.NewError("Google identity token is missing", domainmodelserror.GoogleOAuthInvalid, nil)
	}

	identity, err := o.verifier.Verify(ctx, raw)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "fetching keys") {
			return nil, providerError("Google identity verification failed", err)
		}

		return nil, domainmodelserror.NewError("Google identity token is invalid", domainmodelserror.GoogleOAuthInvalid, nil)
	}

	var claims struct {
		Nonce         string  `json:"nonce"`
		Email         string  `json:"email"`
		EmailVerified bool    `json:"email_verified"`
		Name          *string `json:"name"`
	}

	if err := identity.Claims(&claims); err != nil {
		return nil, domainmodelserror.NewError("Google identity claims are invalid", domainmodelserror.GoogleOAuthInvalid, nil)
	}

	if identity.Subject == "" || claims.Email == "" || !claims.EmailVerified || subtle.ConstantTimeCompare([]byte(o.flow.Hash(claims.Nonce)), []byte(attempt.NonceHash)) != 1 {
		return nil, domainmodelserror.NewError("Google identity is invalid", domainmodelserror.GoogleOAuthInvalid, nil)
	}

	result.Identity = domaincontractsutility.GoogleOAuthIdentity{
		Subject:         identity.Subject,
		Email:           claims.Email,
		IsEmailVerified: claims.EmailVerified,
		Name:            claims.Name,
	}
	return result, nil
}

// Provider SDK errors can contain request URLs and token response bodies.
// Retain only classifications that cannot disclose credentials or provider codes.
func providerError(message string, err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return domainmodelserror.NewError(message, domainmodelserror.Timeout, context.DeadlineExceeded)
	}

	var timeout interface{ Timeout() bool }
	if errors.As(err, &timeout) && timeout.Timeout() {
		return domainmodelserror.NewError(message, domainmodelserror.Timeout, context.DeadlineExceeded)
	}

	if errors.Is(err, context.Canceled) {
		return domainmodelserror.NewError(message, domainmodelserror.BadState, context.Canceled)
	}

	return domainmodelserror.NewError(message, domainmodelserror.GoogleOAuthFailed, nil)
}
