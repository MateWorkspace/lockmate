package applicationsharedtest

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	infrastructurerepositoryclassicauth "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/classic_auth"
	infrastructurerepositoryclient "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/client"
	infrastructurerepositoryrole "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/role"
	infrastructurerepositoryspace "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space"
	infrastructurerepositoryspacemember "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space_member"
	infrastructurerepositoryspacememberrole "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/space_member_role"
	infrastructurerepositoryuser "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/user"
	infrastructureutilitytransactor "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/transactor"
	"github.com/MateWorkspace/mate-golang/pgxdt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresFixture struct {
	Pool            *pgxpool.Pool
	User            domaincontractsrepository.User
	Space           domaincontractsrepository.Space
	Role            domaincontractsrepository.Role
	Client          domaincontractsrepository.Client
	ClassicAuth     domaincontractsrepository.ClassicAuth
	SpaceMember     domaincontractsrepository.SpaceMember
	SpaceMemberRole domaincontractsrepository.SpaceMemberRole
	Transactor      domaincontractsutility.Transactor
}

func NewPostgresFixture(t *testing.T) *PostgresFixture {
	t.Helper()
	url := os.Getenv("LOCKMATE_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("LOCKMATE_TEST_DATABASE_URL is required for isolated PostgreSQL application tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := "application_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext('lockmate/test/extensions')); CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public"); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	_, file, _, _ := runtime.Caller(0)
	migrations := filepath.Join(filepath.Dir(file), "../../../database/migrations")
	paths, err := filepath.Glob(filepath.Join(migrations, "*.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	canonical := regexp.MustCompile(`^\d{14}_[a-z0-9_]+\.up\.sql$`)
	for _, path := range paths {
		if !canonical.MatchString(filepath.Base(path)) {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(data)); err != nil {
			t.Fatalf("migration %s: %v", path, err)
		}
	}
	dt := pgxdt.NewPgxdt(pool)
	question := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
	dollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &PostgresFixture{
		Pool:            pool,
		User:            infrastructurerepositoryuser.NewPostgresImpl(dt, &question, &dollar),
		Space:           infrastructurerepositoryspace.NewPostgresImpl(dt, &question, &dollar),
		Role:            infrastructurerepositoryrole.NewPostgresImpl(dt, &question, &dollar),
		Client:          infrastructurerepositoryclient.NewPostgresImpl(dt, &question, &dollar),
		ClassicAuth:     infrastructurerepositoryclassicauth.NewPostgresImpl(dt, &question, &dollar),
		SpaceMember:     infrastructurerepositoryspacemember.NewPostgresImpl(dt, &question, &dollar),
		SpaceMemberRole: infrastructurerepositoryspacememberrole.NewPostgresImpl(dt, &question, &dollar),
		Transactor:      infrastructureutilitytransactor.NewPgxdtImpl(pgxdt.NewTransactor(pool)),
	}
}
