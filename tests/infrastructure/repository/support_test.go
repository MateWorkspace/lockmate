package repository_test

import (
	"context"
	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	infrastructurerepositoryclassicauth "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/classic_auth"
	infrastructurerepositoryclient "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/client"
	infrastructurerepositorygoogleauth "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/google_auth"
	infrastructurerepositorypermission "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/permission"
	infrastructurerepositoryrole "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/role"
	infrastructurerepositoryrolepermission "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/role_permission"
	infrastructurerepositoryspace "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space"
	infrastructurerepositoryspacemember "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space_member"
	infrastructurerepositoryspacememberrole "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space_member_role"
	infrastructurerepositoryuser "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/user"
	"github.com/MateWorkspace/mate-golang/pgxdt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"testing"
)

type fixture struct {
	ctx                                   context.Context
	pool                                  *pgxpool.Pool
	dt                                    pgxdt.Pgxdt
	spaceId, otherSpaceId, userId, roleId int64
	Space                                 domaincontractsrepository.Space
	Permission                            domaincontractsrepository.Permission
	Role                                  domaincontractsrepository.Role
	User                                  domaincontractsrepository.User
	ClassicAuth                           domaincontractsrepository.ClassicAuth
	GoogleAuth                            domaincontractsrepository.GoogleAuth
	Client                                domaincontractsrepository.Client
	SpaceMember                           domaincontractsrepository.SpaceMember
	RolePermission                        domaincontractsrepository.RolePermission
	SpaceMemberRole                       domaincontractsrepository.SpaceMemberRole
}

func ptr[T any](value T) *T { return &value }
func require(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	url := os.Getenv("LOCKMATE_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("LOCKMATE_TEST_DATABASE_URL is required for isolated PostgreSQL tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	require(t, err)
	schema := "repo_" + regexp.MustCompile("-").ReplaceAllString(uuid.NewString(), "")
	_, err = admin.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public")
	require(t, err)
	_, err = admin.Exec(ctx, "CREATE SCHEMA "+schema)
	require(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		if err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	config, err := pgxpool.ParseConfig(url)
	require(t, err)
	config.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	require(t, err)
	t.Cleanup(pool.Close)
	_, self, _, _ := runtime.Caller(0)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(self), "../../../database/migrations/*.up.sql"))
	require(t, err)
	sort.Strings(files)
	canonical := regexp.MustCompile(`^\d{14}_[a-z_]+\.up\.sql$`)
	for _, file := range files {
		if !canonical.MatchString(filepath.Base(file)) {
			continue
		}
		migration, err := os.ReadFile(file)
		require(t, err)
		_, err = pool.Exec(ctx, string(migration))
		require(t, err)
	}
	dt := pgxdt.NewPgxdt(pool)
	question := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
	dollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	f := &fixture{ctx: ctx, pool: pool, dt: dt}
	f.Space = infrastructurerepositoryspace.NewPostgresImpl(dt, &question, &dollar)
	f.Permission = infrastructurerepositorypermission.NewPostgresImpl(dt, &question, &dollar)
	f.Role = infrastructurerepositoryrole.NewPostgresImpl(dt, &question, &dollar)
	f.User = infrastructurerepositoryuser.NewPostgresImpl(dt, &question, &dollar)
	f.ClassicAuth = infrastructurerepositoryclassicauth.NewPostgresImpl(dt, &question, &dollar)
	f.GoogleAuth = infrastructurerepositorygoogleauth.NewPostgresImpl(dt, &question, &dollar)
	f.Client = infrastructurerepositoryclient.NewPostgresImpl(dt, &question, &dollar)
	f.SpaceMember = infrastructurerepositoryspacemember.NewPostgresImpl(dt, &question, &dollar)
	f.RolePermission = infrastructurerepositoryrolepermission.NewPostgresImpl(dt, &question, &dollar)
	f.SpaceMemberRole = infrastructurerepositoryspacememberrole.NewPostgresImpl(dt, &question, &dollar)
	f.spaceId, err = f.Space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{Slug: "first", Name: "First"})
	require(t, err)
	f.otherSpaceId, err = f.Space.Create(ctx, domaincontractsrepository.CreateSpaceArgs{Slug: "second", Name: "Second"})
	require(t, err)
	f.userId, err = f.User.Create(ctx, domaincontractsrepository.CreateUserArgs{Name: "Fixture User"})
	require(t, err)
	f.roleId, err = f.Role.Create(ctx, f.spaceId, domaincontractsrepository.CreateRoleArgs{Slug: "default", Name: "Default", IsDefault: ptr(true)})
	require(t, err)
	return f
}
