package misc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationmiscsigningkey "github.com/MateWorkspace/lockmate/internal/application/misc/signing_key"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func TestSigningKeyReconcileCreatesRotatesAndRetires(t *testing.T) {
	f := testshared.NewFixture()
	now := time.Now()
	var active *domainmodelsentity.SigningKey
	creates, demotes, retires := 0, 0, 0
	f.SigningKey.ReadActiveFn = func(context.Context) (*domainmodelsentity.SigningKey, error) {
		if active == nil {
			return nil, testshared.NotFound(domainmodelserror.SigningKeyNotFound)
		}
		return active, nil
	}
	f.SigningKey.CreateFn = func(ctx context.Context, value domainmodelsentity.SigningKey) (int64, error) {
		testshared.RequireTx(t, ctx)
		creates++
		v := value
		v.Id = int64(100 + creates)
		active = &v
		return v.Id, nil
	}
	f.SigningKey.DemoteActiveFn = func(ctx context.Context, retiringAt time.Time, by *int64) (int64, error) {
		testshared.RequireTx(t, ctx)
		demotes++
		return active.Id, nil
	}
	f.SigningKey.ListStaleFn = func(context.Context, string) ([]domainmodelsentity.SigningKey, error) { return nil, nil }
	f.SigningKey.DeleteRetiringBeforeFn = func(context.Context, time.Time, *int64) (int64, error) {
		retires++
		return 2, nil
	}
	u := applicationmiscsigningkey.NewUsecaseImpl(f.SigningKey, f.Secret, f.Transactor, f.Logger, applicationmiscsigningkey.Config{RotationInterval: time.Hour, RetentionInterval: 30 * time.Minute})

	if err := u.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if creates != 1 || demotes != 0 {
		t.Fatalf("create reconcile: creates=%d demotes=%d", creates, demotes)
	}

	if err := u.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if creates != 1 || demotes != 0 {
		t.Fatalf("fresh reconcile rotated: creates=%d demotes=%d", creates, demotes)
	}

	active.NotBefore = now.Add(-2 * time.Hour)
	if err := u.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if creates != 2 || demotes != 1 {
		t.Fatalf("stale reconcile did not rotate: creates=%d demotes=%d", creates, demotes)
	}
	if retires == 0 {
		t.Fatal("retention cleanup not invoked")
	}
}

func TestSigningKeyReconcileResealsStaleKeys(t *testing.T) {
	f := testshared.NewFixture()
	updated := 0
	f.SigningKey.ListStaleFn = func(context.Context, string) ([]domainmodelsentity.SigningKey, error) {
		return []domainmodelsentity.SigningKey{{Id: 9, Kid: "old", PrivateKeyCipher: []byte("c"), PrivateKeyNonce: []byte("n"), KeyFingerprint: "old"}}, nil
	}
	f.SigningKey.UpdateSealedFn = func(_ context.Context, id int64, fingerprint string, cipher, nonce []byte) error {
		updated++
		if id != 9 || fingerprint != "current" {
			t.Fatalf("unexpected reseal arguments: %d %q", id, fingerprint)
		}
		return nil
	}
	f.SigningKey.ReadActiveFn = func(context.Context) (*domainmodelsentity.SigningKey, error) {
		return &domainmodelsentity.SigningKey{Id: 80, Kid: "kid", Status: domainmodelsentity.SigningKeyStatusActive, NotBefore: time.Now()}, nil
	}
	f.SigningKey.DeleteRetiringBeforeFn = func(context.Context, time.Time, *int64) (int64, error) { return 0, nil }
	u := applicationmiscsigningkey.NewUsecaseImpl(f.SigningKey, f.Secret, f.Transactor, f.Logger, applicationmiscsigningkey.Config{RotationInterval: time.Hour, RetentionInterval: time.Hour})

	if err := u.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if updated != 1 {
		t.Fatalf("stale key not resealed: %d", updated)
	}
}

func TestSigningKeyReconcilePropagatesFailure(t *testing.T) {
	f := testshared.NewFixture()
	f.SigningKey.ListStaleFn = func(context.Context, string) ([]domainmodelsentity.SigningKey, error) {
		return nil, testshared.NotFound(domainmodelserror.SecretDecryptFailed)
	}
	u := applicationmiscsigningkey.NewUsecaseImpl(f.SigningKey, f.Secret, f.Transactor, f.Logger, applicationmiscsigningkey.Config{RotationInterval: time.Hour, RetentionInterval: time.Hour})

	if err := u.Reconcile(context.Background()); !errors.Is(err, domainmodelserror.SecretDecryptFailed) {
		t.Fatalf("failure not propagated: %v", err)
	}
}
