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

func TestClassicAuthRepository(t *testing.T) {
	f := newFixture(t)
	id, err := f.ClassicAuth.Create(f.ctx, domaincontractsrepository.CreateClassicAuthArgs{UserId: f.userId, Username: "created-user", PasswordHash: "PRIVATE_PASSWORD_HASH", Preferences: ptr(json.RawMessage(`{"created":true}`)), By: ptr(int64(7))})
	require(t, err)
	item, err := f.ClassicAuth.ReadById(f.ctx, id)
	require(t, err)
	if item.Id != id || item.AuditCreate.By == nil || *item.AuditCreate.By != 7 || !json.Valid(item.Preferences) {
		t.Fatalf("bad created entity: %v", item)
	}
	require(t, f.ClassicAuth.UpdateById(f.ctx, id, domaincontractsrepository.UpdateClassicAuthArgs{Username: ptr("updated-user"), Preferences: ptr(json.RawMessage(`{"updated":true}`)), By: ptr(int64(8))}))
	item, err = f.ClassicAuth.ReadById(f.ctx, id)
	require(t, err)
	if item.AuditUpdate.At == nil || item.AuditUpdate.By == nil || *item.AuditUpdate.By != 8 {
		t.Fatal("update audit missing")
	}
	items, total, err := f.ClassicAuth.ReadByFilter(f.ctx, domaincontractsrepository.ClassicAuthFilter{Page: 1, Limit: 10})
	require(t, err)
	if total != 1 || len(items) != 1 {
		t.Fatalf("unexpected page count: %d %d", len(items), total)
	}
	encoded, err := json.Marshal(items)
	require(t, err)
	if strings.Contains(string(encoded), "PRIVATE_") {
		t.Fatal("credential hash leaked in list")
	}
	items, total, err = f.ClassicAuth.ReadByFilter(f.ctx, domaincontractsrepository.ClassicAuthFilter{Page: 999, Limit: 10})
	require(t, err)
	if items == nil || len(items) != 0 || total != 1 {
		t.Fatal("empty page lost total or returned nil slice")
	}
	if _, _, err := f.ClassicAuth.ReadByFilter(f.ctx, domaincontractsrepository.ClassicAuthFilter{Page: math.MaxInt64, Limit: 2}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("pagination overflow accepted")
	}
	require(t, f.ClassicAuth.DeleteById(f.ctx, id, ptr(int64(9))))
	if _, err := f.ClassicAuth.ReadById(f.ctx, id); !errors.Is(err, domainmodelserror.ClassicAuthNotFound) {
		t.Fatalf("deleted entity visible: %v", err)
	}
	if err := f.ClassicAuth.DeleteById(f.ctx, id, nil); !errors.Is(err, domainmodelserror.ClassicAuthNotFound) {
		t.Fatalf("repeat delete accepted: %v", err)
	}
}
