package repository_test

import (
	"context"
	"errors"
	"math"
	"net/netip"
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

func createAllowlistClient(t *testing.T, f *fixture, spaceId int64, slug string) int64 {
	t.Helper()
	id, err := f.Client.Create(f.ctx, spaceId, domaincontractsrepository.CreateClientArgs{
		Slug: slug, Name: "Allowlist Client", SecretKeyHash: "hash", SecretKeyRedacted: "abcd****wxyz",
	})
	require(t, err)

	return id
}

func TestClientIpAllowlistRepository(t *testing.T) {
	f := newFixture(t)
	clientId := createAllowlistClient(t, f, f.spaceId, "web")
	otherClientId := createAllowlistClient(t, f, f.otherSpaceId, "other-web")
	cidr := "203.0.113.0/24"
	id, err := f.ClientIpAllowlist.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{
		ClientId: clientId, Cidr: cidr, By: ptr(int64(7)),
	})
	require(t, err)
	item, err := f.ClientIpAllowlist.ReadById(f.ctx, f.spaceId, id)
	require(t, err)
	if item.ClientId != clientId || item.Cidr != cidr || !item.IsActive || item.AuditCreate.By == nil || *item.AuditCreate.By != 7 {
		t.Fatalf("unexpected created entity: %+v", item)
	}

	byCidr, err := f.ClientIpAllowlist.ReadByClientIdAndCidr(f.ctx, f.spaceId, clientId, cidr)
	require(t, err)
	if byCidr.Id != id {
		t.Fatal("exact CIDR lookup returned a different record")
	}

	_, err = f.ClientIpAllowlist.Create(f.ctx, f.otherSpaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: otherClientId, Cidr: cidr})
	require(t, err)
	secondId, err := f.ClientIpAllowlist.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: "198.51.100.0/24"})
	require(t, err)

	if _, err := f.ClientIpAllowlist.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: cidr}); !errors.Is(err, domainmodelserror.ClientIpAllowlistConflict) {
		t.Fatalf("duplicate create accepted: %v", err)
	}
	if err := f.ClientIpAllowlist.UpdateById(f.ctx, f.spaceId, secondId, domaincontractsrepository.UpdateClientIpAllowlistArgs{Cidr: &cidr}); !errors.Is(err, domainmodelserror.ClientIpAllowlistConflict) {
		t.Fatalf("duplicate update accepted: %v", err)
	}

	require(t, f.ClientIpAllowlist.UpdateById(f.ctx, f.spaceId, id, domaincontractsrepository.UpdateClientIpAllowlistArgs{IsActive: ptr(false), By: ptr(int64(8))}))
	item, err = f.ClientIpAllowlist.ReadById(f.ctx, f.spaceId, id)
	require(t, err)
	if item.IsActive || item.AuditUpdate.At == nil || item.AuditUpdate.By == nil || *item.AuditUpdate.By != 8 || item.ClientId != clientId {
		t.Fatal("update audit, status, or immutable ownership is incorrect")
	}
	if _, err := f.ClientIpAllowlist.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: cidr}); !errors.Is(err, domainmodelserror.ClientIpAllowlistConflict) {
		t.Fatalf("disabled entry released its uniqueness reservation: %v", err)
	}

	items, err := f.ClientIpAllowlist.ReadByClientId(f.ctx, f.spaceId, clientId)
	require(t, err)
	if len(items) != 2 {
		t.Fatal("client list lost an inactive entry")
	}
	for _, filter := range []domaincontractsrepository.ClientIpAllowlistFilter{
		{Page: 1, Limit: 10, ClientId: &clientId, Search: ptr("203.0.113"), IsActive: ptr(false)},
		{Page: 1, Limit: 1, ClientId: &clientId},
	} {
		items, total, err := f.ClientIpAllowlist.ReadByFilter(f.ctx, f.spaceId, filter)
		require(t, err)
		expectedTotal := int64(2)
		if filter.Search != nil {
			expectedTotal = 1
		}
		if len(items) != 1 || total != expectedTotal {
			t.Fatalf("filter/page mismatch: %d rows, %d total", len(items), total)
		}
	}
	for _, filter := range []domaincontractsrepository.ClientIpAllowlistFilter{{Page: 999, Limit: 10}, {Page: 1, Limit: 0}} {
		items, total, err := f.ClientIpAllowlist.ReadByFilter(f.ctx, f.spaceId, filter)
		require(t, err)
		if items == nil || len(items) != 0 || total != 2 {
			t.Fatal("empty page lost count or returned a nil slice")
		}
	}
	if _, _, err := f.ClientIpAllowlist.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.ClientIpAllowlistFilter{Page: math.MaxInt64, Limit: 2}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatalf("pagination overflow accepted: %v", err)
	}

	if _, err := f.ClientIpAllowlist.ReadById(f.ctx, f.otherSpaceId, id); !errors.Is(err, domainmodelserror.ClientIpAllowlistNotFound) {
		t.Fatalf("cross-space read accepted: %v", err)
	}
	if _, err := f.ClientIpAllowlist.ReadByClientIdAndCidr(f.ctx, f.otherSpaceId, clientId, cidr); !errors.Is(err, domainmodelserror.ClientIpAllowlistNotFound) {
		t.Fatalf("cross-space CIDR lookup accepted: %v", err)
	}
	if err := f.ClientIpAllowlist.UpdateById(f.ctx, f.otherSpaceId, id, domaincontractsrepository.UpdateClientIpAllowlistArgs{}); !errors.Is(err, domainmodelserror.ClientIpAllowlistNotFound) {
		t.Fatalf("cross-space update accepted: %v", err)
	}
	if err := f.ClientIpAllowlist.DeleteById(f.ctx, f.otherSpaceId, id, nil); !errors.Is(err, domainmodelserror.ClientIpAllowlistNotFound) {
		t.Fatalf("cross-space delete accepted: %v", err)
	}
	if _, err := f.ClientIpAllowlist.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: otherClientId, Cidr: "192.0.2.0/24"}); !errors.Is(err, domainmodelserror.ClientNotFound) {
		t.Fatalf("cross-space creation accepted: %v", err)
	}
	items, total, err := f.ClientIpAllowlist.ReadByFilter(f.ctx, f.otherSpaceId, domaincontractsrepository.ClientIpAllowlistFilter{ClientId: &clientId, Limit: 10})
	require(t, err)
	if items == nil || len(items) != 0 || total != 0 {
		t.Fatal("cross-space filter exposed records or count")
	}

	items, err = f.ClientIpAllowlist.ReadByClientId(f.ctx, f.otherSpaceId, clientId)
	require(t, err)
	if items == nil || len(items) != 0 {
		t.Fatal("cross-space list exposed records")
	}

	require(t, f.ClientIpAllowlist.DeleteById(f.ctx, f.spaceId, id, ptr(int64(9))))
	var deletedBy int64
	var deletedAt time.Time
	require(t, f.pool.QueryRow(f.ctx, "SELECT deleted_by, deleted_at FROM client_ip_allowlist WHERE id=$1", id).Scan(&deletedBy, &deletedAt))
	if deletedBy != 9 || deletedAt.IsZero() {
		t.Fatal("delete audit missing")
	}
	if _, err := f.ClientIpAllowlist.ReadById(f.ctx, f.spaceId, id); !errors.Is(err, domainmodelserror.ClientIpAllowlistNotFound) {
		t.Fatal("deleted entry visible")
	}
	if err := f.ClientIpAllowlist.DeleteById(f.ctx, f.spaceId, id, nil); !errors.Is(err, domainmodelserror.ClientIpAllowlistNotFound) {
		t.Fatal("repeat deletion succeeded")
	}
	_, err = f.ClientIpAllowlist.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: cidr})
	require(t, err)

	require(t, f.ClientIpAllowlist.UpdateById(f.ctx, f.spaceId, secondId, domaincontractsrepository.UpdateClientIpAllowlistArgs{Cidr: ptr("192.0.2.0/24")}))
	item, err = f.ClientIpAllowlist.ReadById(f.ctx, f.spaceId, secondId)
	require(t, err)
	if item.Cidr != "192.0.2.0/24" {
		t.Fatal("CIDR replacement failed")
	}
}

func TestClientIpAllowlistEligibility(t *testing.T) {
	for _, state := range []string{"active", "entry inactive", "entry deleted", "client inactive", "client deleted", "space inactive", "space deleted"} {
		t.Run(state, func(t *testing.T) {
			f := newFixture(t)
			clientId := createAllowlistClient(t, f, f.spaceId, "web")
			cidr := "203.0.113.0/24"
			ip := netip.MustParseAddr("203.0.113.9")
			id, err := f.ClientIpAllowlist.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: cidr})
			require(t, err)
			switch state {
			case "entry inactive":
				require(t, f.ClientIpAllowlist.UpdateById(f.ctx, f.spaceId, id, domaincontractsrepository.UpdateClientIpAllowlistArgs{IsActive: ptr(false)}))
			case "entry deleted":
				require(t, f.ClientIpAllowlist.DeleteById(f.ctx, f.spaceId, id, nil))
			case "client inactive":
				require(t, f.Client.UpdateById(f.ctx, f.spaceId, clientId, domaincontractsrepository.UpdateClientArgs{IsActive: ptr(false)}))
			case "client deleted":
				require(t, f.Client.DeleteById(f.ctx, f.spaceId, clientId, nil))
			case "space inactive":
				require(t, f.Space.UpdateById(f.ctx, f.spaceId, domaincontractsrepository.UpdateSpaceArgs{IsActive: ptr(false)}))
			case "space deleted":
				require(t, f.Space.DeleteById(f.ctx, f.spaceId, nil))
			}

			scoped, err := f.ClientIpAllowlist.IsAllowedForClient(f.ctx, f.spaceId, clientId, ip)
			require(t, err)
			if scoped != (state == "active") {
				t.Fatalf("incorrect eligibility: scoped=%v", scoped)
			}
			wrongSpace, err := f.ClientIpAllowlist.IsAllowedForClient(f.ctx, f.otherSpaceId, clientId, ip)
			require(t, err)
			if wrongSpace {
				t.Fatal("cross-space allowlist match")
			}
			for _, unknown := range []netip.Addr{netip.MustParseAddr("198.18.0.9"), netip.MustParseAddr("203.0.114.9")} {
				allowed, err := f.ClientIpAllowlist.IsAllowedForClient(f.ctx, f.spaceId, clientId, unknown)
				require(t, err)
				if allowed {
					t.Fatalf("non-member IP matched: %s", unknown)
				}
			}

			_, readErr := f.ClientIpAllowlist.ReadById(f.ctx, f.spaceId, id)
			deleted := state == "entry deleted" || state == "client deleted" || state == "space deleted"
			if deleted {
				items, err := f.ClientIpAllowlist.ReadByClientId(f.ctx, f.spaceId, clientId)
				require(t, err)
				if items == nil || len(items) != 0 {
					t.Fatal("deleted ownership chain exposed list records")
				}
				items, total, err := f.ClientIpAllowlist.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.ClientIpAllowlistFilter{Limit: 10})
				require(t, err)
				if items == nil || len(items) != 0 || total != 0 {
					t.Fatal("deleted ownership chain exposed page records or count")
				}

				if !errors.Is(readErr, domainmodelserror.ClientIpAllowlistNotFound) {
					t.Fatalf("deleted ownership chain exposed: %v", readErr)
				}
			} else {
				require(t, readErr)
			}
			if state == "client deleted" || state == "space deleted" {
				var retained bool
				require(t, f.pool.QueryRow(f.ctx, "SELECT deleted_at IS NULL FROM client_ip_allowlist WHERE id=$1", id).Scan(&retained))
				if !retained {
					t.Fatal("parent deletion changed child audit history")
				}
				if err := f.ClientIpAllowlist.UpdateById(f.ctx, f.spaceId, id, domaincontractsrepository.UpdateClientIpAllowlistArgs{}); !errors.Is(err, domainmodelserror.ClientIpAllowlistNotFound) {
					t.Fatalf("updated entry with deleted parent: %v", err)
				}
				if _, err := f.ClientIpAllowlist.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: "192.0.2.0/24"}); err == nil {
					t.Fatal("created entry with deleted parent")
				}
			}
		})
	}
}

func TestClientIpAllowlistScopedPerClient(t *testing.T) {
	f := newFixture(t)
	clientId := createAllowlistClient(t, f, f.spaceId, "web")
	otherClientId := createAllowlistClient(t, f, f.otherSpaceId, "other-web")
	cidr := "203.0.113.0/24"
	ip := netip.MustParseAddr("203.0.113.9")
	for spaceId, ownerId := range map[int64]int64{f.spaceId: clientId, f.otherSpaceId: otherClientId} {
		_, err := f.ClientIpAllowlist.Create(f.ctx, spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: ownerId, Cidr: cidr})
		require(t, err)
	}
	require(t, f.Client.UpdateById(f.ctx, f.spaceId, clientId, domaincontractsrepository.UpdateClientArgs{IsActive: ptr(false)}))
	scoped, err := f.ClientIpAllowlist.IsAllowedForClient(f.ctx, f.spaceId, clientId, ip)
	require(t, err)
	other, err := f.ClientIpAllowlist.IsAllowedForClient(f.ctx, f.otherSpaceId, otherClientId, ip)
	require(t, err)
	if scoped || !other {
		t.Fatal("client-scoped lookup did not respect independent clients")
	}
}

func TestClientIpAllowlistTransactionRollback(t *testing.T) {
	f := newFixture(t)
	clientId := createAllowlistClient(t, f, f.spaceId, "web")
	cidr := "203.0.113.0/24"
	ip := netip.MustParseAddr("203.0.113.9")
	id, err := f.ClientIpAllowlist.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: cidr})
	require(t, err)
	transactor := pgxdt.NewTransactor(f.pool)
	rollback := errors.New("abort application transaction")
	for _, operation := range []string{"create", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			err := transactor.WithTx(f.ctx, func(txCtx context.Context) error {
				switch operation {
				case "create":
					_, err := f.ClientIpAllowlist.Create(txCtx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: "192.0.2.0/24"})
					if err != nil {
						return err
					}
					allowed, err := f.ClientIpAllowlist.IsAllowedForClient(txCtx, f.spaceId, clientId, netip.MustParseAddr("192.0.2.9"))
					if err != nil {
						return err
					}
					if !allowed {
						t.Fatal("transaction context did not expose its own insert")
					}
				case "update":
					if err := f.ClientIpAllowlist.UpdateById(txCtx, f.spaceId, id, domaincontractsrepository.UpdateClientIpAllowlistArgs{IsActive: ptr(false)}); err != nil {
						return err
					}
				case "delete":
					if err := f.ClientIpAllowlist.DeleteById(txCtx, f.spaceId, id, nil); err != nil {
						return err
					}
				}
				return rollback
			})
			if !errors.Is(err, rollback) {
				t.Fatalf("callback error lost: %v", err)
			}
			allowed, err := f.ClientIpAllowlist.IsAllowedForClient(f.ctx, f.spaceId, clientId, ip)
			require(t, err)
			if !allowed {
				t.Fatal("rollback did not restore the original allowlist")
			}
			items, err := f.ClientIpAllowlist.ReadByClientId(f.ctx, f.spaceId, clientId)
			require(t, err)
			if len(items) != 1 {
				t.Fatal("rolled-back insert was retained")
			}
		})
	}
}

func TestClientIpAllowlistConcurrentDuplicateCreate(t *testing.T) {
	f := newFixture(t)
	clientId := createAllowlistClient(t, f, f.spaceId, "web")
	ctx, cancel := context.WithTimeout(f.ctx, 10*time.Second)
	defer cancel()
	var group sync.WaitGroup
	failures := make(chan error, 4)
	for range 4 {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := f.ClientIpAllowlist.Create(ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: "203.0.113.0/24"})
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
		if !errors.Is(err, domainmodelserror.ClientIpAllowlistConflict) || !errors.As(err, &native) || native.ConstraintName != "uq_client_ip_allowlist_client_id_cidr" {
			t.Fatalf("incorrect duplicate classification or lost native cause: %v", err)
		}
		conflicts++
	}
	if created != 1 || conflicts != 3 {
		t.Fatalf("duplicate invariant failed: created=%d conflicts=%d", created, conflicts)
	}
}

func TestClientIpAllowlistMigrationRoundTrip(t *testing.T) {
	f := newFixture(t)
	clientId := createAllowlistClient(t, f, f.spaceId, "web")
	_, self, _, _ := runtime.Caller(0)
	for _, direction := range []string{"down", "up"} {
		path := filepath.Join(filepath.Dir(self), "../../../database/migrations/20261010025210_create_client_ip_allowlist."+direction+".sql")
		migration, err := os.ReadFile(path)
		require(t, err)
		_, err = f.pool.Exec(f.ctx, string(migration))
		require(t, err)
		var exists bool
		require(t, f.pool.QueryRow(f.ctx, "SELECT to_regclass('client_ip_allowlist') IS NOT NULL").Scan(&exists))
		if exists != (direction == "up") {
			t.Fatalf("unexpected table state after %s", direction)
		}
		_, err = f.Client.ReadById(f.ctx, f.spaceId, clientId)
		require(t, err)
	}
	_, err := f.ClientIpAllowlist.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: "203.0.113.0/24"})
	require(t, err)
}

func TestClientIpAllowlistConcurrentParentDeletion(t *testing.T) {
	for _, parent := range []string{"client", "space"} {
		t.Run(parent, func(t *testing.T) {
			f := newFixture(t)
			clientId := createAllowlistClient(t, f, f.spaceId, "web")
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
					_, err := f.ClientIpAllowlist.Create(ctx, f.spaceId, domaincontractsrepository.CreateClientIpAllowlistArgs{ClientId: clientId, Cidr: "203.0.113.0/24"})
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
			require(t, f.pool.QueryRow(ctx, "SELECT count(*) FROM client_ip_allowlist WHERE client_id=$1", clientId).Scan(&count))
			if count != 0 {
				t.Fatal("entry attached to a deleted parent")
			}
		})
	}
}
