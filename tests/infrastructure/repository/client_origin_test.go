package repository_test

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"github.com/MateWorkspace/mate-golang/pgxdt"
	"github.com/jackc/pgx/v5/pgconn"
)

func createOriginClient(t *testing.T, f *fixture, spaceId int64, slug string) int64 {
	t.Helper()
	id, err := f.Client.Create(f.ctx, spaceId, domaincontractsrepository.CreateClientArgs{
		Slug: slug, Name: "Origin Client", SecretKeyHash: "hash", SecretKeyRedacted: "abcd****wxyz",
	})
	require(t, err)

	return id
}

func TestClientOriginRepository(t *testing.T) {
	f := newFixture(t)
	clientId := createOriginClient(t, f, f.spaceId, "web")
	otherClientId := createOriginClient(t, f, f.otherSpaceId, "other-web")
	origin := "https://app.example.com"
	id, err := f.ClientOrigin.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{
		ClientId: clientId, Origin: origin, By: ptr(int64(7)),
	})
	require(t, err)
	item, err := f.ClientOrigin.ReadById(f.ctx, f.spaceId, id)
	require(t, err)
	if item.ClientId != clientId || item.Origin != origin || !item.IsActive || item.AuditCreate.By == nil || *item.AuditCreate.By != 7 {
		t.Fatalf("unexpected created entity: %+v", item)
	}

	byOrigin, err := f.ClientOrigin.ReadByClientIdAndOrigin(f.ctx, f.spaceId, clientId, origin)
	require(t, err)
	if byOrigin.Id != id {
		t.Fatal("exact origin lookup returned a different record")
	}

	_, err = f.ClientOrigin.Create(f.ctx, f.otherSpaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: otherClientId, Origin: origin})
	require(t, err)
	secondId, err := f.ClientOrigin.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: "https://staging.example.com"})
	require(t, err)

	if _, err := f.ClientOrigin.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: origin}); !errors.Is(err, domainmodelserror.ClientOriginConflict) {
		t.Fatalf("duplicate create accepted: %v", err)
	}
	if err := f.ClientOrigin.UpdateById(f.ctx, f.spaceId, secondId, domaincontractsrepository.UpdateClientOriginArgs{Origin: &origin}); !errors.Is(err, domainmodelserror.ClientOriginConflict) {
		t.Fatalf("duplicate update accepted: %v", err)
	}

	require(t, f.ClientOrigin.UpdateById(f.ctx, f.spaceId, id, domaincontractsrepository.UpdateClientOriginArgs{IsActive: ptr(false), By: ptr(int64(8))}))
	item, err = f.ClientOrigin.ReadById(f.ctx, f.spaceId, id)
	require(t, err)
	if item.IsActive || item.AuditUpdate.At == nil || item.AuditUpdate.By == nil || *item.AuditUpdate.By != 8 || item.ClientId != clientId {
		t.Fatal("update audit, status, or immutable ownership is incorrect")
	}
	if _, err := f.ClientOrigin.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: origin}); !errors.Is(err, domainmodelserror.ClientOriginConflict) {
		t.Fatalf("disabled origin released its uniqueness reservation: %v", err)
	}

	items, err := f.ClientOrigin.ReadByClientId(f.ctx, f.spaceId, clientId)
	require(t, err)
	if len(items) != 2 {
		t.Fatal("client list lost an inactive origin")
	}
	for _, filter := range []domaincontractsrepository.ClientOriginFilter{
		{Page: 1, Limit: 10, ClientId: &clientId, Search: ptr("app.example"), IsActive: ptr(false)},
		{Page: 1, Limit: 1, ClientId: &clientId},
	} {
		items, total, err := f.ClientOrigin.ReadByFilter(f.ctx, f.spaceId, filter)
		require(t, err)
		expectedTotal := int64(2)
		if filter.Search != nil {
			expectedTotal = 1
		}
		if len(items) != 1 || total != expectedTotal {
			t.Fatalf("filter/page mismatch: %d rows, %d total", len(items), total)
		}
	}
	for _, filter := range []domaincontractsrepository.ClientOriginFilter{{Page: 999, Limit: 10}, {Page: 1, Limit: 0}} {
		items, total, err := f.ClientOrigin.ReadByFilter(f.ctx, f.spaceId, filter)
		require(t, err)
		if items == nil || len(items) != 0 || total != 2 {
			t.Fatal("empty page lost count or returned a nil slice")
		}
	}
	if _, _, err := f.ClientOrigin.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.ClientOriginFilter{Page: math.MaxInt64, Limit: 2}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatalf("pagination overflow accepted: %v", err)
	}

	if _, err := f.ClientOrigin.ReadById(f.ctx, f.otherSpaceId, id); !errors.Is(err, domainmodelserror.ClientOriginNotFound) {
		t.Fatalf("cross-space read accepted: %v", err)
	}
	if _, err := f.ClientOrigin.ReadByClientIdAndOrigin(f.ctx, f.otherSpaceId, clientId, origin); !errors.Is(err, domainmodelserror.ClientOriginNotFound) {
		t.Fatalf("cross-space origin lookup accepted: %v", err)
	}
	if err := f.ClientOrigin.UpdateById(f.ctx, f.otherSpaceId, id, domaincontractsrepository.UpdateClientOriginArgs{}); !errors.Is(err, domainmodelserror.ClientOriginNotFound) {
		t.Fatalf("cross-space update accepted: %v", err)
	}
	if err := f.ClientOrigin.DeleteById(f.ctx, f.otherSpaceId, id, nil); !errors.Is(err, domainmodelserror.ClientOriginNotFound) {
		t.Fatalf("cross-space delete accepted: %v", err)
	}
	if _, err := f.ClientOrigin.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: otherClientId, Origin: "https://other.example.com"}); !errors.Is(err, domainmodelserror.ClientNotFound) {
		t.Fatalf("cross-space creation accepted: %v", err)
	}
	items, total, err := f.ClientOrigin.ReadByFilter(f.ctx, f.otherSpaceId, domaincontractsrepository.ClientOriginFilter{ClientId: &clientId, Limit: 10})
	require(t, err)
	if items == nil || len(items) != 0 || total != 0 {
		t.Fatal("cross-space filter exposed records or count")
	}

	items, err = f.ClientOrigin.ReadByClientId(f.ctx, f.otherSpaceId, clientId)
	require(t, err)
	if items == nil || len(items) != 0 {
		t.Fatal("cross-space list exposed records")
	}

	require(t, f.ClientOrigin.DeleteById(f.ctx, f.spaceId, id, ptr(int64(9))))
	var deletedBy int64
	var deletedAt time.Time
	require(t, f.pool.QueryRow(f.ctx, "SELECT deleted_by, deleted_at FROM client_origins WHERE id=$1", id).Scan(&deletedBy, &deletedAt))
	if deletedBy != 9 || deletedAt.IsZero() {
		t.Fatal("delete audit missing")
	}
	if _, err := f.ClientOrigin.ReadById(f.ctx, f.spaceId, id); !errors.Is(err, domainmodelserror.ClientOriginNotFound) {
		t.Fatal("deleted origin visible")
	}
	if err := f.ClientOrigin.DeleteById(f.ctx, f.spaceId, id, nil); !errors.Is(err, domainmodelserror.ClientOriginNotFound) {
		t.Fatal("repeat deletion succeeded")
	}
	_, err = f.ClientOrigin.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: origin})
	require(t, err)

	require(t, f.ClientOrigin.UpdateById(f.ctx, f.spaceId, secondId, domaincontractsrepository.UpdateClientOriginArgs{Origin: ptr("https://preview.example.com")}))
	item, err = f.ClientOrigin.ReadById(f.ctx, f.spaceId, secondId)
	require(t, err)
	if item.Origin != "https://preview.example.com" {
		t.Fatal("origin replacement failed")
	}
}

func TestClientOriginEligibility(t *testing.T) {
	for _, state := range []string{"active", "origin inactive", "origin deleted", "client inactive", "client deleted", "space inactive", "space deleted"} {
		t.Run(state, func(t *testing.T) {
			f := newFixture(t)
			clientId := createOriginClient(t, f, f.spaceId, "web")
			origin := "https://app.example.com"
			id, err := f.ClientOrigin.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: origin})
			require(t, err)
			switch state {
			case "origin inactive":
				require(t, f.ClientOrigin.UpdateById(f.ctx, f.spaceId, id, domaincontractsrepository.UpdateClientOriginArgs{IsActive: ptr(false)}))
			case "origin deleted":
				require(t, f.ClientOrigin.DeleteById(f.ctx, f.spaceId, id, nil))
			case "client inactive":
				require(t, f.Client.UpdateById(f.ctx, f.spaceId, clientId, domaincontractsrepository.UpdateClientArgs{IsActive: ptr(false)}))
			case "client deleted":
				require(t, f.Client.DeleteById(f.ctx, f.spaceId, clientId, nil))
			case "space inactive":
				require(t, f.Space.UpdateById(f.ctx, f.spaceId, domaincontractsrepository.UpdateSpaceArgs{IsActive: ptr(false)}))
			case "space deleted":
				require(t, f.Space.DeleteById(f.ctx, f.spaceId, nil))
			}

			global, err := f.ClientOrigin.IsAllowedOrigin(f.ctx, origin)
			require(t, err)
			scoped, err := f.ClientOrigin.IsAllowedOriginForClient(f.ctx, f.spaceId, clientId, origin)
			require(t, err)
			if global != (state == "active") || scoped != global {
				t.Fatalf("incorrect eligibility: global=%v scoped=%v", global, scoped)
			}
			wrongSpace, err := f.ClientOrigin.IsAllowedOriginForClient(f.ctx, f.otherSpaceId, clientId, origin)
			require(t, err)
			if wrongSpace {
				t.Fatal("cross-space allowlist match")
			}
			for _, unknown := range []string{"https://unknown.example.com", "http://app.example.com", "https://app.example.com:8443", "https://app.example.com.evil.test", "null", ""} {
				allowed, err := f.ClientOrigin.IsAllowedOrigin(f.ctx, unknown)
				require(t, err)
				if allowed {
					t.Fatalf("non-exact origin matched: %s", unknown)
				}
			}

			_, readErr := f.ClientOrigin.ReadById(f.ctx, f.spaceId, id)
			deleted := state == "origin deleted" || state == "client deleted" || state == "space deleted"
			if deleted {
				items, err := f.ClientOrigin.ReadByClientId(f.ctx, f.spaceId, clientId)
				require(t, err)
				if items == nil || len(items) != 0 {
					t.Fatal("deleted ownership chain exposed list records")
				}
				items, total, err := f.ClientOrigin.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.ClientOriginFilter{Limit: 10})
				require(t, err)
				if items == nil || len(items) != 0 || total != 0 {
					t.Fatal("deleted ownership chain exposed page records or count")
				}

				if !errors.Is(readErr, domainmodelserror.ClientOriginNotFound) {
					t.Fatalf("deleted ownership chain exposed: %v", readErr)
				}
			} else {
				require(t, readErr)
			}
			if state == "client deleted" || state == "space deleted" {
				var retained bool
				require(t, f.pool.QueryRow(f.ctx, "SELECT deleted_at IS NULL FROM client_origins WHERE id=$1", id).Scan(&retained))
				if !retained {
					t.Fatal("parent deletion changed child audit history")
				}
				if err := f.ClientOrigin.UpdateById(f.ctx, f.spaceId, id, domaincontractsrepository.UpdateClientOriginArgs{}); !errors.Is(err, domainmodelserror.ClientOriginNotFound) {
					t.Fatalf("updated origin with deleted parent: %v", err)
				}
				if _, err := f.ClientOrigin.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: "https://new.example.com"}); err == nil {
					t.Fatal("created origin with deleted parent")
				}
			}
		})
	}
}

func TestClientOriginGlobalLookupAcrossClients(t *testing.T) {
	f := newFixture(t)
	clientId := createOriginClient(t, f, f.spaceId, "web")
	otherClientId := createOriginClient(t, f, f.otherSpaceId, "other-web")
	origin := "https://shared.example.com"
	for spaceId, ownerId := range map[int64]int64{f.spaceId: clientId, f.otherSpaceId: otherClientId} {
		_, err := f.ClientOrigin.Create(f.ctx, spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: ownerId, Origin: origin})
		require(t, err)
	}
	require(t, f.Client.UpdateById(f.ctx, f.spaceId, clientId, domaincontractsrepository.UpdateClientArgs{IsActive: ptr(false)}))
	global, err := f.ClientOrigin.IsAllowedOrigin(f.ctx, origin)
	require(t, err)
	scoped, err := f.ClientOrigin.IsAllowedOriginForClient(f.ctx, f.spaceId, clientId, origin)
	require(t, err)
	if !global || scoped {
		t.Fatal("global lookup did not preserve the other client's independent allowance")
	}

	require(t, f.Client.DeleteById(f.ctx, f.otherSpaceId, otherClientId, nil))
	global, err = f.ClientOrigin.IsAllowedOrigin(f.ctx, origin)
	require(t, err)
	if global {
		t.Fatal("global origin remained allowed with no eligible owners")
	}
}

func TestClientOriginTransactionRollback(t *testing.T) {
	f := newFixture(t)
	clientId := createOriginClient(t, f, f.spaceId, "web")
	origin := "https://app.example.com"
	id, err := f.ClientOrigin.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: origin})
	require(t, err)
	transactor := pgxdt.NewTransactor(f.pool)
	rollback := errors.New("abort application transaction")
	for _, operation := range []string{"create", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			err := transactor.WithTx(f.ctx, func(txCtx context.Context) error {
				switch operation {
				case "create":
					_, err := f.ClientOrigin.Create(txCtx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: "https://new.example.com"})
					if err != nil {
						return err
					}
					allowed, err := f.ClientOrigin.IsAllowedOriginForClient(txCtx, f.spaceId, clientId, "https://new.example.com")
					if err != nil {
						return err
					}
					if !allowed {
						t.Fatal("transaction context did not expose its own insert")
					}
				case "update":
					if err := f.ClientOrigin.UpdateById(txCtx, f.spaceId, id, domaincontractsrepository.UpdateClientOriginArgs{IsActive: ptr(false)}); err != nil {
						return err
					}
				case "delete":
					if err := f.ClientOrigin.DeleteById(txCtx, f.spaceId, id, nil); err != nil {
						return err
					}
				}
				return rollback
			})
			if !errors.Is(err, rollback) {
				t.Fatalf("callback error lost: %v", err)
			}
			allowed, err := f.ClientOrigin.IsAllowedOrigin(f.ctx, origin)
			require(t, err)
			if !allowed {
				t.Fatal("rollback did not restore the original allowlist")
			}
			items, err := f.ClientOrigin.ReadByClientId(f.ctx, f.spaceId, clientId)
			require(t, err)
			if len(items) != 1 {
				t.Fatal("rolled-back insert was retained")
			}
		})
	}
}

func TestClientOriginConcurrentDuplicateCreate(t *testing.T) {
	f := newFixture(t)
	clientId := createOriginClient(t, f, f.spaceId, "web")
	ctx, cancel := context.WithTimeout(f.ctx, 10*time.Second)
	defer cancel()
	var group sync.WaitGroup
	failures := make(chan error, 4)
	for range 4 {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := f.ClientOrigin.Create(ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: "https://app.example.com"})
			failures <- err
		}()
	}
	group.Wait()
	close(failures)
	created, conflicts := 0, 0
	for err := range failures {
		if err == nil {
			created++
			continue
		}
		var native *pgconn.PgError
		if !errors.Is(err, domainmodelserror.ClientOriginConflict) || !errors.As(err, &native) || native.ConstraintName != "uq_client_origins_client_id_origin" {
			t.Fatalf("incorrect duplicate classification or lost native cause: %v", err)
		}
		conflicts++
	}
	if created != 1 || conflicts != 3 {
		t.Fatalf("duplicate invariant failed: created=%d conflicts=%d", created, conflicts)
	}
}

func TestClientOriginMigrationRoundTrip(t *testing.T) {
	f := newFixture(t)
	clientId := createOriginClient(t, f, f.spaceId, "web")
	_, self, _, _ := runtime.Caller(0)
	for _, direction := range []string{"down", "up"} {
		path := filepath.Join(filepath.Dir(self), "../../../database/migrations/20261010025210_create_client_origins."+direction+".sql")
		migration, err := os.ReadFile(path)
		require(t, err)
		_, err = f.pool.Exec(f.ctx, string(migration))
		require(t, err)
		var exists bool
		require(t, f.pool.QueryRow(f.ctx, "SELECT to_regclass('client_origins') IS NOT NULL").Scan(&exists))
		if exists != (direction == "up") {
			t.Fatalf("unexpected table state after %s", direction)
		}
		_, err = f.Client.ReadById(f.ctx, f.spaceId, clientId)
		require(t, err)
	}
	_, err := f.ClientOrigin.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: "https://app.example.com"})
	require(t, err)
}

func TestClientOriginConcurrentParentDeletion(t *testing.T) {
	for _, parent := range []string{"client", "space"} {
		t.Run(parent, func(t *testing.T) {
			f := newFixture(t)
			clientId := createOriginClient(t, f, f.spaceId, "web")
			ctx, cancel := context.WithTimeout(f.ctx, 5*time.Second)
			defer cancel()
			completed := make(chan error, 1)
			transactor := pgxdt.NewTransactor(f.pool)
			require(t, transactor.WithTx(ctx, func(txCtx context.Context) error {
				if parent == "client" {
					if err := f.Client.DeleteById(txCtx, f.spaceId, clientId, nil); err != nil {
						return err
					}
				} else if err := f.Space.DeleteById(txCtx, f.spaceId, nil); err != nil {
					return err
				}

				started := make(chan struct{})
				go func() {
					close(started)
					_, err := f.ClientOrigin.Create(ctx, f.spaceId, domaincontractsrepository.CreateClientOriginArgs{ClientId: clientId, Origin: "https://app.example.com"})
					completed <- err
				}()
				<-started
				return nil
			}))

			kind := domainmodelserror.ClientNotFound
			if parent == "space" {
				kind = domainmodelserror.SpaceNotFound
			}
			if err := <-completed; !errors.Is(err, kind) {
				t.Fatalf("creation survived parent deletion or lost classification: %v", err)
			}
			var count int64
			require(t, f.pool.QueryRow(ctx, "SELECT count(*) FROM client_origins WHERE client_id=$1", clientId).Scan(&count))
			if count != 0 {
				t.Fatal("origin attached to a deleted parent")
			}
		})
	}
}
