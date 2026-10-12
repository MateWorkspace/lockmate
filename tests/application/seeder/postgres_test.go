package seeder_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	applicationauthclient "github.com/MateWorkspace/lockmate/internal/application/auth/client"
	applicationauthsession "github.com/MateWorkspace/lockmate/internal/application/auth/session"
	applicationmanagementspace "github.com/MateWorkspace/lockmate/internal/application/management/space"
	applicationseeder "github.com/MateWorkspace/lockmate/internal/application/seeder"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainmodelsseeder "github.com/MateWorkspace/lockmate/internal/domain/models/seeder"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	domainusecasesseeder "github.com/MateWorkspace/lockmate/internal/domain/usecases/seeder"
	infrastructurerepositorypermission "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/permission"
	infrastructurerepositoryrolepermission "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/role_permission"
	infrastructureutilitykeygen "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/keygen"
	infrastructureutilitylogger "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/logger"
	infrastructureutilitypassword "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/password"
	infrastructureutilitypermissionguard "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/permission_guard"
	infrastructureutilitytoken "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/token"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

type seedFixture struct {
	*testshared.PostgresFixture
	data           domainmodelsseeder.Data
	permission     domaincontractsrepository.Permission
	rolePermission domaincontractsrepository.RolePermission
	logger         *testshared.Logger
	password       domaincontractsutility.Password
	keygen         domaincontractsutility.Keygen
	validator      domaincontractsutility.Validator
}

func seedDatabase(t *testing.T) *seedFixture {
	t.Helper()
	base := testshared.NewPostgresFixture(t)
	dt := pgxdt.NewPgxdt(base.Pool)
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
	d := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &seedFixture{PostgresFixture: base, data: definitions(t), permission: infrastructurerepositorypermission.NewPostgresImpl(dt, &q, &d), rolePermission: infrastructurerepositoryrolepermission.NewPostgresImpl(dt, &q, &d), logger: &testshared.Logger{}, password: infrastructureutilitypassword.NewBcryptImpl(4), keygen: infrastructureutilitykeygen.NewGeneratorImpl(), validator: infrastructureutilityvalidator.NewValidatorImpl(true)}
}
func (f *seedFixture) seeder(logger domaincontractsutility.Logger) domainusecasesseeder.Seeder {
	return applicationseeder.NewUsecaseImpl(f.Space, f.permission, f.Role, f.rolePermission, f.User, f.ClassicAuth, f.SpaceMember, f.SpaceMemberRole, f.Client, f.ClientIpAllowlist, f.password, f.keygen, f.validator, f.Transactor, logger, &f.data)
}
func (f *seedFixture) count(t *testing.T, table string) int {
	t.Helper()
	var count int
	if err := f.Pool.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
func (f *seedFixture) assertBaseline(t *testing.T) {
	t.Helper()
	for table, expected := range map[string]int{"spaces": 1, "permissions": 46, "roles": 3, "role_permission": 83, "users": 3, "classic_auths": 3, "space_members": 3, "space_member_role": 5, "clients": 1, "client_ip_allowlist": 2} {
		if actual := f.count(t, table); actual != expected {
			t.Fatalf("%s: got %d, expected %d", table, actual, expected)
		}
	}
}

func TestPostgresSeedBaselineAndPreservation(t *testing.T) {
	f := seedDatabase(t)
	ctx := context.Background()
	u := f.seeder(f.logger)
	if err := u.Run(ctx); err != nil {
		t.Fatal(err)
	}
	f.assertBaseline(t)
	initialLogs := len(f.logger.Entries)
	if initialLogs == 0 {
		t.Fatal("missing committed creation logs")
	}
	if err := u.Run(ctx); err != nil {
		t.Fatal(err)
	}
	f.assertBaseline(t)
	if len(f.logger.Entries) != initialLogs {
		t.Fatal("idempotent seed produced mutation logs")
	}
	if strings.Contains(fmt.Sprint(f.logger.Entries), f.data.Clients[0].SecretKey) {
		t.Fatal("seed secret logged")
	}
	for _, user := range f.data.Users {
		if strings.Contains(fmt.Sprint(f.logger.Entries), user.Password) {
			t.Fatal("seed password logged")
		}
	}
	client, err := f.Client.ReadBySlug(ctx, "lockmate")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.keygen.Compare(client.SecretKeyHash, f.data.Clients[0].SecretKey); err != nil {
		t.Fatal(err)
	}
	if len(client.SecretKeyRedacted) != len(f.data.Clients[0].SecretKey) || client.AuthRedirectUri != nil {
		t.Fatal("client redaction/default callback")
	}
	auth, err := f.ClassicAuth.ReadByUsername(ctx, "super")
	if err != nil {
		t.Fatal(err)
	}
	changedPassword, err := f.password.Hash("changed-password")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Pool.Exec(ctx, "UPDATE users SET name='Edited Super' WHERE id=$1;", auth.UserId); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Pool.Exec(ctx, "UPDATE classic_auths SET password_hash=$1 WHERE id=$2", changedPassword, auth.Id); err != nil {
		t.Fatal(err)
	}
	otherKey, otherHash, otherRedaction, err := f.keygen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	uri := "https://project.example.com/callback"
	if err := f.Client.UpdateById(ctx, client.SpaceId, client.Id, domaincontractsrepository.UpdateClientArgs{SecretKeyHash: &otherHash, SecretKeyRedacted: &otherRedaction, AuthRedirectUri: &uri}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Pool.Exec(ctx, "UPDATE roles SET name='Edited Role' WHERE space_id=$1 AND slug='super'", client.SpaceId); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Pool.Exec(ctx, "UPDATE space_members SET is_active=false WHERE user_id=$1", auth.UserId); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Pool.Exec(ctx, "UPDATE client_ip_allowlist SET is_active=false"); err != nil {
		t.Fatal(err)
	}
	if err := u.Run(ctx); err != nil {
		t.Fatal(err)
	}
	preserved, err := f.ClassicAuth.ReadByUsername(ctx, "super")
	if err != nil {
		t.Fatal(err)
	}
	if preserved.PasswordHash != changedPassword {
		t.Fatal("seed reset password")
	}
	profile, err := f.User.ReadById(ctx, auth.UserId)
	if err != nil || profile.Name != "Edited Super" {
		t.Fatal("seed overwrote profile")
	}
	client, err = f.Client.ReadBySlug(ctx, "lockmate")
	if err != nil || client.AuthRedirectUri == nil || *client.AuthRedirectUri != uri || f.keygen.Compare(client.SecretKeyHash, otherKey) != nil {
		t.Fatal("seed reset client settings")
	}
	member, err := f.SpaceMember.ReadByUserId(ctx, client.SpaceId, auth.UserId)
	if err != nil || member.IsActive {
		t.Fatal("seed reactivated membership")
	}
	entries, err := f.ClientIpAllowlist.ReadByClientId(ctx, client.SpaceId, client.Id)
	if err != nil || entries[0].IsActive {
		t.Fatal("seed reactivated IP allowlist entry")
	}
	role, err := f.Role.ReadBySlug(ctx, client.SpaceId, "super")
	if err != nil || role.Name != "Edited Role" {
		t.Fatal("seed overwrote role")
	}
}

func TestPostgresSeedRestoresMissingGrants(t *testing.T) {
	f := seedDatabase(t)
	ctx := context.Background()
	u := f.seeder(f.logger)
	if err := u.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Pool.Exec(ctx, "DELETE FROM role_permission WHERE id=(SELECT min(id) FROM role_permission); DELETE FROM space_member_role WHERE id=(SELECT max(id) FROM space_member_role)"); err != nil {
		t.Fatal(err)
	}
	if err := u.Run(ctx); err != nil {
		t.Fatal(err)
	}
	f.assertBaseline(t)
}

func TestPostgresSeedRollbackAndConcurrentStartup(t *testing.T) {
	t.Run("late write failure", func(t *testing.T) {
		f := seedDatabase(t)
		if _, err := f.Pool.Exec(context.Background(), "ALTER TABLE clients ADD CONSTRAINT reject_seed_client CHECK (slug <> 'lockmate')"); err != nil {
			t.Fatal(err)
		}
		if err := f.seeder(f.logger).Run(context.Background()); err == nil {
			t.Fatal("seed unexpectedly succeeded")
		}
		for _, table := range []string{"spaces", "permissions", "roles", "users", "classic_auths", "space_members", "clients"} {
			if f.count(t, table) != 0 {
				t.Fatal("partial seed retained")
			}
		}
		for _, entry := range f.logger.Entries {
			if entry.Level == "info" {
				t.Fatal("uncommitted success log")
			}
		}
	})
	t.Run("commit failure", func(t *testing.T) {
		f := seedDatabase(t)
		sql := `CREATE FUNCTION reject_seed_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'seed commit rejected'; RETURN NEW; END $$;
   CREATE CONSTRAINT TRIGGER reject_seed_commit AFTER INSERT ON clients DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_seed_commit();`
		if _, err := f.Pool.Exec(context.Background(), sql); err != nil {
			t.Fatal(err)
		}
		if err := f.seeder(f.logger).Run(context.Background()); err == nil {
			t.Fatal("failed commit accepted")
		}
		if f.count(t, "spaces") != 0 || f.count(t, "users") != 0 {
			t.Fatal("failed commit retained seed")
		}
		for _, entry := range f.logger.Entries {
			if entry.Level == "info" {
				t.Fatal("failed commit success log")
			}
		}
	})
	t.Run("concurrent", func(t *testing.T) {
		f := seedDatabase(t)
		logger := infrastructureutilitylogger.NewSlogImpl(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, domainmodelslogger.LevelNone, false)
		var wg sync.WaitGroup
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := f.seeder(logger).Run(context.Background()); err != nil {
					t.Error(err)
				}
			}()
		}
		wg.Wait()
		f.assertBaseline(t)
	})
}

func TestPostgresSeedOwnershipAndDefaultRole(t *testing.T) {
	t.Run("other client owner", func(t *testing.T) {
		f := seedDatabase(t)
		ctx := context.Background()
		spaceId, err := f.Space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{Slug: "other", Name: "Other"})
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.Client.Create(ctx, spaceId, domaincontractsrepository.CreateClientArgs{Slug: "lockmate", Name: "Other Client", SecretKeyHash: "hash", SecretKeyRedacted: "masked"})
		if err != nil {
			t.Fatal(err)
		}
		if err := f.seeder(f.logger).Run(ctx); !errors.Is(err, domainmodelserror.BadState) {
			t.Fatal(err)
		}
		if f.count(t, "spaces") != 1 || f.count(t, "permissions") != 0 {
			t.Fatal("ownership failure retained writes")
		}
	})
	t.Run("preserve custom default", func(t *testing.T) {
		f := seedDatabase(t)
		ctx := context.Background()
		spaceId, err := f.Space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{Slug: "lockmate", Name: "Existing"})
		if err != nil {
			t.Fatal(err)
		}
		yes := true
		customId, err := f.Role.Create(ctx, spaceId, domaincontractsrepository.CreateRoleArgs{Slug: "custom", Name: "Custom", IsDefault: &yes})
		if err != nil {
			t.Fatal(err)
		}
		if err := f.seeder(f.logger).Run(ctx); err != nil {
			t.Fatal(err)
		}
		role, err := f.Role.ReadDefault(ctx, spaceId)
		if err != nil || role.Id != customId {
			t.Fatal("seed replaced default")
		}
	})
	t.Run("missing existing default", func(t *testing.T) {
		f := seedDatabase(t)
		ctx := context.Background()
		spaceId, err := f.Space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{Slug: "lockmate", Name: "Existing"})
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.Role.Create(ctx, spaceId, domaincontractsrepository.CreateRoleArgs{Slug: "user", Name: "Existing User"})
		if err != nil {
			t.Fatal(err)
		}
		if err := f.seeder(f.logger).Run(ctx); !errors.Is(err, domainmodelserror.BadState) {
			t.Fatal(err)
		}
		if f.count(t, "permissions") != 0 {
			t.Fatal("missing default retained writes")
		}
	})
}

func TestPostgresSeededAuthenticationAndAuthority(t *testing.T) {
	f := seedDatabase(t)
	ctx := context.Background()
	f.data.Space.Slug = "custom-control"
	if err := f.seeder(f.logger).Run(ctx); err != nil {
		t.Fatal(err)
	}
	clientUsecase := applicationauthclient.NewUsecaseImpl(f.Client, f.Space, f.keygen, f.logger)
	client, err := clientUsecase.Authenticate(ctx, domainusecasesauth.AuthenticateClientRequest{Slug: "lockmate", Secret: f.data.Clients[0].SecretKey})
	if err != nil {
		t.Fatal(err)
	}
	token := infrastructureutilitytoken.NewJwtImpl("test-refresh", time.Minute, time.Hour)
	session := applicationauthsession.NewUsecaseImpl(f.Client, f.Space, f.User, f.ClassicAuth, f.SpaceMember, f.SpaceMemberRole, f.SigningKey, f.password, token, f.Secret, f.validator, f.Transactor, f.logger)
	guard := infrastructureutilitypermissionguard.NewClaimsImpl()
	management := applicationmanagementspace.NewUsecaseImpl(f.User, f.Space, f.SpaceMember, guard, f.validator, f.logger, f.Transactor, f.data.Space.Slug)
	for _, user := range f.data.Users {
		result, err := session.SignIn(ctx, client, domainusecasesauth.SignInRequest{Username: user.Username, Password: user.Password})
		if err != nil {
			t.Fatal(err)
		}
		if result.AccessToken == "" || result.RefreshToken == "" {
			t.Fatal("seeded session incomplete")
		}
		claims, err := session.AuthenticateAccessToken(ctx, client, result.AccessToken)
		if err != nil {
			t.Fatal(err)
		}
		if err := guard.Check(claims, []string{"profile:get", "profile_security:get"}); err != nil {
			t.Fatal(err)
		}
		_, err = management.ReadByFilter(ctx, claims, domainusecasesmanagement.FilterSpaceRequest{Page: 1, Limit: 10})
		if user.Username == "user" {
			if !errors.Is(err, domainmodelserror.Forbidden) {
				t.Fatal("user received administration")
			}
		} else if err != nil {
			t.Fatal(err)
		}
		_, err = management.Create(ctx, claims, domainusecasesmanagement.CreateSpaceRequest{Slug: "project-" + user.Username, Name: "Project"})
		if user.Username == "super" {
			if err != nil {
				t.Fatal(err)
			}
		} else if !errors.Is(err, domainmodelserror.Forbidden) {
			t.Fatal("non-super changed space catalog")
		}
	}
}
