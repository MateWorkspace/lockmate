package repository_test

import (
	"errors"
	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructurerepositoryauthhandoff "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/auth_handoff"
	infrastructurerepositorygoogleoauthattempt "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/google_oauth_attempt"
	infrastructurerepositorygoogleoauthlaunch "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/google_oauth_launch"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAuthFlowPersistence(t *testing.T) {
	f := newFixture(t)
	redirect := "https://app.example.com/callback"
	clientId, err := f.Client.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientArgs{Slug: "oauth", Name: "OAuth", SecretKeyHash: "hash", SecretKeyRedacted: "masked", AuthRedirectUri: &redirect})
	require(t, err)
	item, err := f.Client.ReadById(f.ctx, f.spaceId, clientId)
	require(t, err)
	if item.AuthRedirectUri == nil || *item.AuthRedirectUri != redirect {
		t.Fatal("redirect missing")
	}
	summaries, _, err := f.Client.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.ClientFilter{Page: 1, Limit: 10})
	require(t, err)
	if summaries[0].AuthRedirectUri == nil {
		t.Fatal("redirect missing in projection")
	}
	if err := f.Client.UpdateById(f.ctx, f.spaceId, clientId, domaincontractsrepository.UpdateClientArgs{AuthRedirectUri: &redirect, SetAuthRedirectUriNull: true}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal(err)
	}
	require(t, f.Client.UpdateById(f.ctx, f.spaceId, clientId, domaincontractsrepository.UpdateClientArgs{SetAuthRedirectUriNull: true}))
	item, err = f.Client.ReadById(f.ctx, f.spaceId, clientId)
	require(t, err)
	if item.AuthRedirectUri != nil {
		t.Fatal("redirect not cleared")
	}
	d := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	launches := infrastructurerepositorygoogleoauthlaunch.NewPostgresImpl(f.dt, &d)
	attempts := infrastructurerepositorygoogleoauthattempt.NewPostgresImpl(f.dt, &d)
	handoffs := infrastructurerepositoryauthhandoff.NewPostgresImpl(f.dt, &d)
	launch := domainmodelsentity.GoogleOAuthLaunch{TicketHash: "launch-digest", ClientId: clientId, SpaceId: f.spaceId, RedirectUri: redirect, ClientState: "state", ExpiresAt: time.Now().Add(time.Minute)}
	require(t, launches.Create(f.ctx, launch))
	var wg sync.WaitGroup
	var successes atomic.Int64
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := launches.Consume(f.ctx, launch.TicketHash)
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, domainmodelserror.GoogleOAuthStateInvalid) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatal("launch was not single-use")
	}
	attempt := domainmodelsentity.GoogleOAuthAttempt{StateHash: "state-digest", BindingHash: "binding-digest", NonceHash: "nonce-digest", Verifier: "pkce-secret", ClientId: clientId, SpaceId: f.spaceId, RedirectUri: redirect, ClientState: "state", ExpiresAt: time.Now().Add(time.Minute)}
	require(t, attempts.Create(f.ctx, attempt))
	if _, err := attempts.Consume(f.ctx, attempt.StateHash, "wrong-binding"); !errors.Is(err, domainmodelserror.GoogleOAuthStateInvalid) {
		t.Fatal(err)
	}
	consumed, err := attempts.Consume(f.ctx, attempt.StateHash, attempt.BindingHash)
	require(t, err)
	if consumed.Verifier != attempt.Verifier {
		t.Fatal("PKCE verifier lost")
	}
	if _, err := attempts.Consume(f.ctx, attempt.StateHash, attempt.BindingHash); !errors.Is(err, domainmodelserror.GoogleOAuthStateInvalid) {
		t.Fatal(err)
	}
	launch.TicketHash = "expired"
	launch.ExpiresAt = time.Now().Add(-time.Second)
	require(t, launches.Create(f.ctx, launch))
	if _, err := launches.Consume(f.ctx, launch.TicketHash); !errors.Is(err, domainmodelserror.GoogleOAuthStateInvalid) {
		t.Fatal(err)
	}
	attempt.StateHash = "expired"
	attempt.ExpiresAt = time.Now().Add(-time.Second)
	require(t, attempts.Create(f.ctx, attempt))
	if _, err := attempts.Consume(f.ctx, attempt.StateHash, attempt.BindingHash); !errors.Is(err, domainmodelserror.GoogleOAuthStateInvalid) {
		t.Fatal(err)
	}
	memberId, err := f.SpaceMember.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: f.userId})
	require(t, err)
	require(t, handoffs.Create(f.ctx, domainmodelsentity.AuthHandoff{CodeHash: "expired", ClientId: clientId, SpaceId: f.spaceId, UserId: f.userId, SpaceMemberId: memberId, RedirectUri: redirect, ExpiresAt: time.Now().Add(-time.Second)}))
	for _, purge := range []func() (int64, error){func() (int64, error) { return launches.PurgeExpired(f.ctx, 1000) }, func() (int64, error) { return attempts.PurgeExpired(f.ctx, 1000) }, func() (int64, error) { return handoffs.PurgeExpired(f.ctx, 1000) }} {
		n, err := purge()
		require(t, err)
		if n != 1 {
			t.Fatalf("cleanup deleted %d rows", n)
		}
	}
	if _, err := launches.PurgeExpired(f.ctx, 1001); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal(err)
	}
}

func TestAuthFlowMigrationRollback(t *testing.T) {
	f := newFixture(t)
	_, self, _, _ := runtime.Caller(0)
	directory := filepath.Join(filepath.Dir(self), "../../../database/migrations")
	names := []string{"20261012000100_add_client_auth_redirect_uri", "20261012000200_create_google_oauth_launches", "20261012000300_create_google_oauth_attempts", "20261012000400_create_auth_handoffs"}
	for i := len(names) - 1; i >= 0; i-- {
		sql, err := os.ReadFile(filepath.Join(directory, names[i]+".down.sql"))
		require(t, err)
		_, err = f.pool.Exec(f.ctx, string(sql))
		require(t, err)
	}
	for _, name := range names {
		sql, err := os.ReadFile(filepath.Join(directory, name+".up.sql"))
		require(t, err)
		_, err = f.pool.Exec(f.ctx, string(sql))
		require(t, err)
	}
}
