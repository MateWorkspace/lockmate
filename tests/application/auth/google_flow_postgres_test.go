package auth_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	applicationauthgoogle "github.com/MateWorkspace/lockmate/internal/application/auth/google"
	applicationauthhandoff "github.com/MateWorkspace/lockmate/internal/application/auth/handoff"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	infrastructurerepositoryauthhandoff "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/auth_handoff"
	infrastructurerepositorygoogleauth "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/google_auth"
	infrastructurerepositorygoogleoauthlaunch "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/google_oauth_launch"
	infrastructureutilityauthflow "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/auth_flow"
	infrastructureutilitylogger "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/logger"
	infrastructureutilitytoken "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/token"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

func TestPostgresGoogleHandoffLifecycle(t *testing.T) {
	f := testshared.NewPostgresFixture(t)
	ctx := context.Background()
	spaceId, err := f.Space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{Slug: "google-space", Name: "Google Space"})
	if err != nil {
		t.Fatal(err)
	}
	yes := true
	if _, err = f.Role.Create(ctx, spaceId, domaincontractsrepository.CreateRoleArgs{Slug: "member", Name: "Member", IsDefault: &yes}); err != nil {
		t.Fatal(err)
	}
	redirect := "https://project.example.com/auth/callback"
	clientId, err := f.Client.Create(ctx, spaceId, domaincontractsrepository.CreateClientArgs{Slug: "google-client", Name: "Google Client", SecretKeyHash: "hash", SecretKeyRedacted: "masked", AuthRedirectUri: &redirect})
	if err != nil {
		t.Fatal(err)
	}
	client := &domainmodelsentity.ClientIdentity{Id: clientId, SpaceId: spaceId, Slug: "google-client"}
	dt := pgxdt.NewPgxdt(f.Pool)
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
	d := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	identities := infrastructurerepositorygoogleauth.NewPostgresImpl(dt, &q, &d)
	launches := infrastructurerepositorygoogleoauthlaunch.NewPostgresImpl(dt, &d)
	handoffs := infrastructurerepositoryauthhandoff.NewPostgresImpl(dt, &d)
	flow := infrastructureutilityauthflow.NewRandomImpl()
	logger := infrastructureutilitylogger.NewSlogImpl(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, domainmodelslogger.LevelNone, false)
	provider := &testshared.GoogleOAuth{CompleteFn: func(context.Context, domaincontractsutility.CompleteGoogleOAuthArgs) (*domaincontractsutility.CompleteGoogleOAuthResult, error) {
		return &domaincontractsutility.CompleteGoogleOAuthResult{ClientId: clientId, SpaceId: spaceId, RedirectUri: redirect, ClientState: "project-state", Identity: domaincontractsutility.GoogleOAuthIdentity{Subject: "stable-google-sub", Email: "user@example.com", IsEmailVerified: true}}, nil
	}}
	google := applicationauthgoogle.NewUsecaseImpl(f.Client, f.Space, f.User, f.SpaceMember, f.SpaceMemberRole, infrastructureutilityvalidator.NewValidatorImpl(true), f.Transactor, logger, identities, provider, launches, handoffs, flow, applicationauthgoogle.Config{Enabled: true, PublicUrl: "https://lockmate-dev.dhonan.my.id", LaunchDuration: time.Minute, AttemptDuration: time.Minute, HandoffDuration: time.Minute})
	token := infrastructureutilitytoken.NewJwtImpl("test-access", "test-refresh", time.Minute, time.Hour)
	makeExchange := func(token domaincontractsutility.Token) domainusecasesauth.Handoff {
		return applicationauthhandoff.NewUsecaseImpl(handoffs, f.Client, f.Space, f.User, f.SpaceMember, f.SpaceMemberRole, flow, token, f.Transactor, logger)
	}
	exchange := makeExchange(token)
	callback := func() string {
		result, err := google.SignInCallback(ctx, callbackRequest())
		if err != nil {
			t.Fatal(err)
		}
		uri, err := url.Parse(result.RedirectUrl)
		if err != nil {
			t.Fatal(err)
		}
		if uri.Query().Get("state") != "project-state" || len(uri.Query()) != 2 {
			t.Fatal("unsafe redirect")
		}
		return uri.Query().Get("code")
	}
	first := callback()
	var identitiesCount int
	if err := f.Pool.QueryRow(ctx, "SELECT count(*) FROM google_auths").Scan(&identitiesCount); err != nil || identitiesCount != 1 {
		t.Fatalf("registration not committed: %v", err)
	}
	// Another valid client cannot redeem the original client's code.
	otherId, err := f.Client.Create(ctx, spaceId, domaincontractsrepository.CreateClientArgs{Slug: "other-client", Name: "Other Client", SecretKeyHash: "hash", SecretKeyRedacted: "masked", AuthRedirectUri: &redirect})
	if err != nil {
		t.Fatal(err)
	}
	_, err = exchange.Exchange(ctx, &domainmodelsentity.ClientIdentity{Id: otherId, SpaceId: spaceId, Slug: "other-client"}, domainusecasesauth.ExchangeAuthHandoffRequest{Code: first})
	if !errors.Is(err, domainmodelserror.AuthHandoffInvalid) {
		t.Fatalf("wrong client exchanged code: %v", err)
	}
	result, err := exchange.Exchange(ctx, client, domainusecasesauth.ExchangeAuthHandoffRequest{Code: first})
	if err != nil || result == nil || !result.IsNewUser || result.AccessToken == "" || result.RefreshToken == "" {
		t.Fatalf("exchange failed: %v", err)
	}
	if _, err := exchange.Exchange(ctx, client, domainusecasesauth.ExchangeAuthHandoffRequest{Code: first}); !errors.Is(err, domainmodelserror.AuthHandoffInvalid) {
		t.Fatalf("replay accepted: %v", err)
	}
	// A token generation failure rolls back consumption; a subsequent exchange succeeds.
	retry := callback()
	failedToken := &testshared.Token{GenerateAccessFn: func(*domainmodelsentity.AccessClaims) (string, error) {
		return "", domainmodelserror.NewError("token failure", domainmodelserror.TokenFailed, nil)
	}}
	if _, err := makeExchange(failedToken).Exchange(ctx, client, domainusecasesauth.ExchangeAuthHandoffRequest{Code: retry}); !errors.Is(err, domainmodelserror.TokenFailed) {
		t.Fatal(err)
	}
	if _, err := exchange.Exchange(ctx, client, domainusecasesauth.ExchangeAuthHandoffRequest{Code: retry}); err != nil {
		t.Fatal(err)
	}
	// Configuration changes invalidate outstanding handoffs without redirecting elsewhere.
	changed := callback()
	replacement := "https://project.example.com/another-callback"
	if err := f.Client.UpdateById(ctx, spaceId, clientId, domaincontractsrepository.UpdateClientArgs{AuthRedirectUri: &replacement}); err != nil {
		t.Fatal(err)
	}
	if _, err := exchange.Exchange(ctx, client, domainusecasesauth.ExchangeAuthHandoffRequest{Code: changed}); !errors.Is(err, domainmodelserror.BadState) {
		t.Fatal(err)
	}
	if err := f.Client.UpdateById(ctx, spaceId, clientId, domaincontractsrepository.UpdateClientArgs{AuthRedirectUri: &redirect}); err != nil {
		t.Fatal(err)
	}
	// Suspended memberships cannot exchange a saved code.
	blocked := callback()
	if _, err := f.Pool.Exec(ctx, "UPDATE space_members SET is_active=false WHERE id=$1", result.SpaceMember.Id); err != nil {
		t.Fatal(err)
	}
	if _, err := exchange.Exchange(ctx, client, domainusecasesauth.ExchangeAuthHandoffRequest{Code: blocked}); !errors.Is(err, domainmodelserror.Unauthorized) {
		t.Fatal(err)
	}
	if _, err := f.Pool.Exec(ctx, "UPDATE space_members SET is_active=true WHERE id=$1", result.SpaceMember.Id); err != nil {
		t.Fatal(err)
	}
	// Exactly one concurrent exchange commits.
	concurrent := callback()
	var successes atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := exchange.Exchange(ctx, client, domainusecasesauth.ExchangeAuthHandoffRequest{Code: concurrent})
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, domainmodelserror.AuthHandoffInvalid) {
				t.Errorf("unexpected concurrent error: %v", err)
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("%d exchanges committed", successes.Load())
	}
	expired, err := flow.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if err := handoffs.Create(ctx, domainmodelsentity.AuthHandoff{CodeHash: flow.Hash(expired), ClientId: clientId, SpaceId: spaceId, UserId: result.User.Id, SpaceMemberId: result.SpaceMember.Id, RedirectUri: redirect, ExpiresAt: time.Now().Add(-time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err := exchange.Exchange(ctx, client, domainusecasesauth.ExchangeAuthHandoffRequest{Code: expired}); !errors.Is(err, domainmodelserror.AuthHandoffExpired) {
		t.Fatal(err)
	}
	// Later callbacks reuse the global identity and never overwrite profile edits.
	if _, err := f.Pool.Exec(ctx, "UPDATE users SET name='Edited Name' WHERE id=$1", result.User.Id); err != nil {
		t.Fatal(err)
	}
	last := callback()
	again, err := exchange.Exchange(ctx, client, domainusecasesauth.ExchangeAuthHandoffRequest{Code: last})
	if err != nil || again.IsNewUser || again.User.Name != "Edited Name" {
		t.Fatalf("existing identity changed: %v", err)
	}
}
