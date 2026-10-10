package misc_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	applicationmiscallowedorigin "github.com/MateWorkspace/lockmate/internal/application/misc/allowed_origin"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func allowedOrigin(f *testshared.Fixture) domainusecasesmisc.AllowedOrigin {
	return applicationmiscallowedorigin.NewUsecaseImpl(f.ClientOrigin, f.Validator, f.Logger)
}

func checkOrigin(u domainusecasesmisc.AllowedOrigin, ctx context.Context, scoped bool, origin string) (bool, error) {
	if scoped {
		return u.IsAllowedOriginForClient(ctx, 10, 30, origin)
	}
	return u.IsAllowedOrigin(ctx, origin)
}

func TestAllowedOriginDelegationAndQuietNonmatches(t *testing.T) {
	for _, scoped := range []bool{false, true} {
		for _, expected := range []bool{false, true} {
			t.Run(map[bool]string{false: "global", true: "client"}[scoped]+map[bool]string{false: "/denied", true: "/allowed"}[expected], func(t *testing.T) {
				f := testshared.NewFixture()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				called := false
				if scoped {
					f.ClientOrigin.IsAllowedOriginForClientFn = func(actual context.Context, spaceId, clientId int64, origin string) (bool, error) {
						called = true
						if actual != ctx || spaceId != 10 || clientId != 30 || origin != "https://app.example.com" {
							t.Fatal("client check lost context, scope, or origin")
						}
						return expected, nil
					}
				} else {
					f.ClientOrigin.IsAllowedOriginFn = func(actual context.Context, origin string) (bool, error) {
						called = true
						if actual != ctx || origin != "https://app.example.com" {
							t.Fatal("global check lost context or origin")
						}
						return expected, nil
					}
				}
				actual, err := checkOrigin(allowedOrigin(f), ctx, scoped, "https://app.example.com")
				if err != nil || actual != expected || !called || len(f.Logger.Entries) != 0 || f.Transactions != 0 {
					t.Fatalf("unexpected check result: %v %v", actual, err)
				}
			})
		}
	}
}

func TestAllowedOriginInvalidInputShortCircuits(t *testing.T) {
	for _, scoped := range []bool{false, true} {
		for _, origin := range []string{"", "null", "https://*.example.com", "https://app.example.com/", "http://example.com", strings.Repeat("a", 2049)} {
			f := testshared.NewFixture()
			allowed, err := checkOrigin(allowedOrigin(f), context.Background(), scoped, origin)
			if allowed || err != nil || len(f.Logger.Entries) != 0 {
				t.Fatalf("invalid origin was not a quiet nonmatch: %v %v", allowed, err)
			}
		}
	}
}

func TestAllowedOriginOperationalFailures(t *testing.T) {
	for _, scoped := range []bool{false, true} {
		for _, validation := range []bool{false, true} {
			for _, source := range []error{errors.New("database unavailable"), context.Canceled, context.DeadlineExceeded} {
				f := testshared.NewFixture()
				failure := domainmodelserror.NewError("check failed", domainmodelserror.Failure, source)
				if validation {
					f.Validator = &testshared.Validator{ClientOriginFn: func(string) error { return failure }}
				} else if scoped {
					f.ClientOrigin.IsAllowedOriginForClientFn = func(context.Context, int64, int64, string) (bool, error) { return true, failure }
				} else {
					f.ClientOrigin.IsAllowedOriginFn = func(context.Context, string) (bool, error) { return true, failure }
				}
				allowed, err := checkOrigin(allowedOrigin(f), context.Background(), scoped, "https://app.example.com")
				if allowed || err != failure || !errors.Is(err, source) {
					t.Fatalf("error became a nonmatch or cause lost: %v %v", allowed, err)
				}
				if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "error" || f.Logger.Entries[0].Meta["err"] != failure {
					t.Fatal("operational failure omitted its explicit error log")
				}
				if validation && !strings.Contains(f.Logger.Entries[0].Message, "validate") || !validation && !strings.Contains(f.Logger.Entries[0].Message, "check") {
					t.Fatal("error message did not identify the failed step")
				}
			}
		}
	}
}

func TestAllowedOriginInvalidIdentifiers(t *testing.T) {
	for _, ids := range [][2]int64{{0, 30}, {10, 0}, {-1, 30}} {
		f := testshared.NewFixture()
		f.Validator = &testshared.Validator{}
		allowed, err := allowedOrigin(f).IsAllowedOriginForClient(context.Background(), ids[0], ids[1], "https://app.example.com")
		testshared.RequireError(t, err, domainmodelserror.BadArgs)
		if allowed || len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Meta["space_id"] != ids[0] || f.Logger.Entries[0].Meta["client_id"] != ids[1] {
			t.Fatal("invalid scope reached validation or omitted diagnostic IDs")
		}
	}
}
