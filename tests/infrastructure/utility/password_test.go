package utility_test

import (
	"errors"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitypassword "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/password"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"testing"
)

func TestPassword(t *testing.T) {
	password := infrastructureutilitypassword.NewBcryptImpl(bcrypt.MinCost)
	hash, err := password.Hash("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	if err := password.Compare(hash, "correct-password"); err != nil {
		t.Fatal(err)
	}
	if err := password.Compare(hash, "wrong-password"); !errors.Is(err, domainmodelserror.PasswordMismatch) || !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("wrong mismatch classification: %v", err)
	}
	if !errors.Is(password.Compare("invalid-hash", "password"), domainmodelserror.PasswordCompareFailed) {
		t.Fatal("malformed hash accepted")
	}
	if _, err := password.Hash(strings.Repeat("a", 73)); !errors.Is(err, domainmodelserror.PasswordHashFailed) || !errors.Is(err, bcrypt.ErrPasswordTooLong) {
		t.Fatalf("wrong hash failure: %v", err)
	}
	if !errors.Is(password.Compare(hash, strings.Repeat("a", 73)), domainmodelserror.PasswordCompareFailed) {
		t.Fatal("oversized password accepted")
	}
}

func TestPasswordDefaultCost(t *testing.T) {
	hash, err := infrastructureutilitypassword.NewBcryptImpl(-1).Hash("password")
	if err != nil {
		t.Fatal(err)
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil || cost != bcrypt.DefaultCost {
		t.Fatalf("unexpected cost: %d %v", cost, err)
	}
}
