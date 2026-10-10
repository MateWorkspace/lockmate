package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationauthsession "github.com/MateWorkspace/lockmate/internal/application/auth/session"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitypassword "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/password"
	infrastructureutilitytoken "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/token"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

func TestPostgresRegistrationCommitBoundary(t *testing.T) {
	f := testshared.NewPostgresFixture(t)
	ctx := context.Background()
	spaceId, err := f.Space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{Slug: "lockmate", Name: "Lockmate"})
	if err != nil {
		t.Fatal(err)
	}
	yes := true
	roleId, err := f.Role.Create(ctx, spaceId, domaincontractsrepository.CreateRoleArgs{Slug: "user", Name: "User", IsDefault: &yes})
	if err != nil {
		t.Fatal(err)
	}
	clientId, err := f.Client.Create(ctx, spaceId, domaincontractsrepository.CreateClientArgs{Slug: "web", Name: "Web client", SecretKeyHash: "hash", SecretKeyRedacted: "masked"})
	if err != nil {
		t.Fatal(err)
	}
	client := &domainmodelsapp.ClientIdentity{Id: clientId, SpaceId: spaceId, Slug: "web"}
	password := infrastructureutilitypassword.NewBcryptImpl(4)
	token := infrastructureutilitytoken.NewJwtImpl("test-access-secret", "test-refresh-secret", time.Minute, time.Hour)
	validator := infrastructureutilityvalidator.NewValidatorImpl()
	logger := &testshared.Logger{}
	u := applicationauthsession.NewUsecaseImpl(f.Client, f.Space, f.User, f.ClassicAuth, f.SpaceMember, f.SpaceMemberRole, password, token, validator, f.Transactor, logger)
	result, err := u.Register(ctx, client, registerRequest())
	if err != nil {
		t.Fatal(err)
	}
	if result.AccessToken == "" || result.RefreshToken == "" || len(result.Roles) != 1 || result.Roles[0].Id != roleId {
		t.Fatal("missing tokens or atomic default-role assignment")
	}
	if _, err := token.ValidateAccess(spaceId, result.AccessToken); err != nil {
		t.Fatal(err)
	}
	before := 0
	if err := f.Pool.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&before); err != nil {
		t.Fatal(err)
	}
	fail := &testshared.Token{GenerateAccessFn: func(*domainmodelsentity.AccessClaims) (string, error) {
		return "", testshared.NotFound(domainmodelserror.TokenFailed)
	}}
	failing := applicationauthsession.NewUsecaseImpl(f.Client, f.Space, f.User, f.ClassicAuth, f.SpaceMember, f.SpaceMemberRole, password, fail, validator, f.Transactor, logger)
	request := registerRequest()
	request.Username = "failed-user"
	failed, err := failing.Register(ctx, client, request)
	testshared.RequireError(t, err, domainmodelserror.TokenFailed)
	if failed != nil {
		t.Fatal("failed transaction returned credentials")
	}
	after := 0
	if err := f.Pool.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("token failure retained registration writes")
	}
	err = f.Transactor.WithTx(ctx, func(txCtx context.Context) error {
		request.Username = "nested-user"
		nested, nestedErr := u.Register(txCtx, client, request)
		if nested != nil || !errors.Is(nestedErr, domainmodelserror.BadState) || !errors.Is(nestedErr, pgxdt.ErrTransactionAlreadyActive) {
			t.Fatalf("nested feature not rejected: %v", nestedErr)
		}
		return nestedErr
	})
	testshared.RequireError(t, err, domainmodelserror.BadState)
	if err := f.Role.DeleteById(ctx, spaceId, roleId, nil); err != nil {
		t.Fatal(err)
	}
	request.Username = "no-default"
	failed, err = u.Register(ctx, client, request)
	testshared.RequireError(t, err, domainmodelserror.DefaultRoleNotFound)
	if failed != nil {
		t.Fatal("missing role returned registration")
	}
	if err := f.Pool.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("missing default role retained user")
	}
}
