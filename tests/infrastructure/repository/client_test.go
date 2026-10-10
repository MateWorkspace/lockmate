package repository_test

import (
	"encoding/json"
	"errors"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"math"
	"strings"
	"testing"
)

func TestClientRepository(t *testing.T) {
	f := newFixture(t)
	id, err := f.Client.Create(f.ctx, f.spaceId, domaincontractsrepository.CreateClientArgs{Slug: "created", Name: "Created Name", SecretKeyHash: "PRIVATE_SECRET_HASH", SecretKeyRedacted: "abcd****wxyz", Preferences: ptr(json.RawMessage(`{"created":true}`)), By: ptr(int64(7))})
	require(t, err)
	item, err := f.Client.ReadById(f.ctx, f.spaceId, id)
	require(t, err)
	if item.Id != id || item.AuditCreate.By == nil || *item.AuditCreate.By != 7 || !json.Valid(item.Preferences) {
		t.Fatalf("bad created entity: %v", item)
	}
	require(t, f.Client.UpdateById(f.ctx, f.spaceId, id, domaincontractsrepository.UpdateClientArgs{Name: ptr("Updated Name"), Preferences: ptr(json.RawMessage(`{"updated":true}`)), By: ptr(int64(8))}))
	item, err = f.Client.ReadById(f.ctx, f.spaceId, id)
	require(t, err)
	if item.AuditUpdate.At == nil || item.AuditUpdate.By == nil || *item.AuditUpdate.By != 8 {
		t.Fatal("update audit missing")
	}
	items, total, err := f.Client.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.ClientFilter{Page: 1, Limit: 10})
	require(t, err)
	if total != 1 || len(items) != 1 {
		t.Fatalf("unexpected page count: %d %d", len(items), total)
	}
	encoded, err := json.Marshal(items)
	require(t, err)
	if strings.Contains(string(encoded), "PRIVATE_") {
		t.Fatal("credential hash leaked in list")
	}
	items, total, err = f.Client.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.ClientFilter{Page: 999, Limit: 10})
	require(t, err)
	if items == nil || len(items) != 0 || total != 1 {
		t.Fatal("empty page lost total or returned nil slice")
	}
	if _, _, err := f.Client.ReadByFilter(f.ctx, f.spaceId, domaincontractsrepository.ClientFilter{Page: math.MaxInt64, Limit: 2}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("pagination overflow accepted")
	}
	if _, err := f.Client.ReadById(f.ctx, f.otherSpaceId, id); !errors.Is(err, domainmodelserror.ClientNotFound) {
		t.Fatalf("cross-space read accepted: %v", err)
	}
	if err := f.Client.UpdateById(f.ctx, f.otherSpaceId, id, domaincontractsrepository.UpdateClientArgs{}); !errors.Is(err, domainmodelserror.ClientNotFound) {
		t.Fatalf("cross-space update accepted: %v", err)
	}
	require(t, f.Client.DeleteById(f.ctx, f.spaceId, id, ptr(int64(9))))
	if _, err := f.Client.ReadById(f.ctx, f.spaceId, id); !errors.Is(err, domainmodelserror.ClientNotFound) {
		t.Fatalf("deleted entity visible: %v", err)
	}
	if err := f.Client.DeleteById(f.ctx, f.spaceId, id, nil); !errors.Is(err, domainmodelserror.ClientNotFound) {
		t.Fatalf("repeat delete accepted: %v", err)
	}
}
