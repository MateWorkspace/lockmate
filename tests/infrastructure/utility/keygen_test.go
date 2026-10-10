package utility_test

import (
	"errors"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitykeygen "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/keygen"
	"strings"
	"testing"
)

func TestKeygen(t *testing.T) {
	generator := infrastructureutilitykeygen.NewGeneratorImpl()
	seen := map[string]bool{}
	for range 100 {
		raw, hash, redacted, err := generator.Generate()
		if err != nil {
			t.Fatal(err)
		}
		if len(raw) != len("lockmate-")+64 || len(hash) != 64 || !strings.HasPrefix(raw, "lockmate-") || hash != generator.Hash(raw) || seen[raw] {
			t.Fatal("invalid or duplicate generated key")
		}
		seen[raw] = true
		if redacted != raw[:4]+strings.Repeat("*", len(raw)-8)+raw[len(raw)-4:] {
			t.Fatal("incorrect redaction")
		}
		if err := generator.Compare(hash, raw); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(generator.Compare(strings.Repeat("0", 64), raw), domainmodelserror.KeyMismatch) {
			t.Fatal("mismatched key accepted")
		}
	}
	for _, invalid := range []string{"", "lockmate-", "wrongxxx" + strings.Repeat("a", 64), "lockmate-" + strings.Repeat("A", 64), "lockmate-" + strings.Repeat("g", 64)} {
		if !errors.Is(generator.Validate(invalid), domainmodelserror.KeyInvalid) {
			t.Fatalf("accepted malformed key %q", invalid)
		}
	}
}
