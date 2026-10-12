package misc_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	applicationmiscipallowlist "github.com/MateWorkspace/lockmate/internal/application/misc/ip_allowlist"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func ipAllowlist(f *testshared.Fixture) domainusecasesmisc.IpAllowlist {
	return applicationmiscipallowlist.NewUsecaseImpl(f.ClientIpAllowlist, f.Logger)
}

func TestIpAllowlistDelegationAndQuietNonmatches(t *testing.T) {
	for _, expected := range []bool{false, true} {
		t.Run(map[bool]string{false: "denied", true: "allowed"}[expected], func(t *testing.T) {
			f := testshared.NewFixture()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			called := false
			f.ClientIpAllowlist.IsAllowedForClientFn = func(actual context.Context, spaceId, clientId int64, ip netip.Addr) (bool, error) {
				called = true
				if actual != ctx || spaceId != 10 || clientId != 30 || ip.String() != "203.0.113.9" {
					t.Fatal("check lost context, scope, or IP")
				}
				return expected, nil
			}
			actual, err := ipAllowlist(f).IsAllowedIpForClient(ctx, 10, 30, "203.0.113.9")
			if err != nil || actual != expected || !called || len(f.Logger.Entries) != 0 || f.Transactions != 0 {
				t.Fatalf("unexpected check result: %v %v", actual, err)
			}
		})
	}
}

func TestIpAllowlistLoopbackAndInvalidInputs(t *testing.T) {
	f := testshared.NewFixture()
	f.ClientIpAllowlist.IsAllowedForClientFn = func(context.Context, int64, int64, netip.Addr) (bool, error) {
		t.Fatal("loopback or invalid input reached the repository")
		return false, nil
	}
	for _, ip := range []string{"127.0.0.1", "::1"} {
		allowed, err := ipAllowlist(f).IsAllowedIpForClient(context.Background(), 10, 30, ip)
		if err != nil || !allowed {
			t.Fatalf("loopback not allowed: %v %v", allowed, err)
		}
	}
	for _, ip := range []string{"", "not-an-ip", "999.1.1.1"} {
		allowed, err := ipAllowlist(f).IsAllowedIpForClient(context.Background(), 10, 30, ip)
		if allowed || err != nil {
			t.Fatalf("invalid IP was not a quiet nonmatch: %v %v", allowed, err)
		}
	}
	if len(f.Logger.Entries) != 0 {
		t.Fatal("quiet paths logged")
	}
}

func TestIpAllowlistOperationalFailures(t *testing.T) {
	for _, source := range []error{errors.New("database unavailable"), context.Canceled, context.DeadlineExceeded} {
		f := testshared.NewFixture()
		failure := domainmodelserror.NewError("check failed", domainmodelserror.Failure, source)
		f.ClientIpAllowlist.IsAllowedForClientFn = func(context.Context, int64, int64, netip.Addr) (bool, error) { return true, failure }
		allowed, err := ipAllowlist(f).IsAllowedIpForClient(context.Background(), 10, 30, "203.0.113.9")
		if allowed || err != failure || !errors.Is(err, source) {
			t.Fatalf("error became a nonmatch or cause lost: %v %v", allowed, err)
		}
		if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Level != "error" || f.Logger.Entries[0].Meta["err"] != failure {
			t.Fatal("operational failure omitted its explicit error log")
		}
	}
}

func TestIpAllowlistInvalidIdentifiers(t *testing.T) {
	for _, ids := range [][2]int64{{0, 30}, {10, 0}, {-1, 30}} {
		f := testshared.NewFixture()
		f.ClientIpAllowlist.IsAllowedForClientFn = func(context.Context, int64, int64, netip.Addr) (bool, error) {
			t.Fatal("invalid scope reached the repository")
			return false, nil
		}
		allowed, err := ipAllowlist(f).IsAllowedIpForClient(context.Background(), ids[0], ids[1], "203.0.113.9")
		testshared.RequireError(t, err, domainmodelserror.BadArgs)
		if allowed || len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Meta["space_id"] != ids[0] || f.Logger.Entries[0].Meta["client_id"] != ids[1] {
			t.Fatal("invalid scope omitted diagnostic IDs")
		}
	}
}
