package utility_test

import (
	"errors"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitytoken "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/token"
	"github.com/golang-jwt/jwt/v5"
	"reflect"
	"testing"
	"time"
)

func TestTokenRoundTripAndPurpose(t *testing.T) {
	token := infrastructureutilitytoken.NewJwtImpl("same-secret", "same-secret", time.Minute, time.Hour)
	want := &domainmodelsentity.AccessClaims{UserId: 1, SpaceId: 2, SpaceMemberId: 3, Name: "Google User", Roles: []string{"user"}, Permissions: []string{"profile:get"}}
	access, err := token.GenerateAccess(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := token.ValidateAccess(2, access)
	if err != nil || !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected claims: %v %v", got, err)
	}
	refreshClaims := &domainmodelsentity.RefreshClaims{UserId: 1, SpaceId: 2, SpaceMemberId: 3, Name: "Google User"}
	refresh, err := token.GenerateRefresh(refreshClaims)
	if err != nil {
		t.Fatal(err)
	}
	gotRefresh, err := token.ValidateRefresh(2, refresh)
	if err != nil || !reflect.DeepEqual(refreshClaims, gotRefresh) {
		t.Fatalf("unexpected refresh: %v %v", gotRefresh, err)
	}
	if _, err := token.ValidateAccess(9, access); !errors.Is(err, domainmodelserror.TokenInvalid) {
		t.Fatal("wrong space accepted")
	}
	if _, err := token.ValidateAccess(2, refresh); !errors.Is(err, domainmodelserror.TokenInvalid) {
		t.Fatal("refresh used as access")
	}
	if _, err := token.ValidateRefresh(2, access); !errors.Is(err, domainmodelserror.TokenInvalid) {
		t.Fatal("access used as refresh")
	}
}

func TestTokenValidationFailures(t *testing.T) {
	token := infrastructureutilitytoken.NewJwtImpl("secret", "refresh", time.Minute, time.Hour)
	for _, test := range []struct {
		name     string
		change   func(jwt.MapClaims)
		method   jwt.SigningMethod
		secret   string
		expected error
	}{
		{"expired", func(c jwt.MapClaims) {
			c["iat"] = time.Now().Add(-2 * time.Hour).Unix()
			c["nbf"] = c["iat"]
			c["exp"] = time.Now().Add(-time.Hour).Unix()
		}, jwt.SigningMethodHS256, "secret", domainmodelserror.TokenExpired},
		{"signature", func(c jwt.MapClaims) {}, jwt.SigningMethodHS256, "wrong", domainmodelserror.TokenInvalid},
		{"algorithm", func(c jwt.MapClaims) {}, jwt.SigningMethodHS512, "secret", domainmodelserror.TokenInvalid},
		{"no expiration", func(c jwt.MapClaims) { delete(c, "exp") }, jwt.SigningMethodHS256, "secret", domainmodelserror.TokenInvalid},
		{"no issued at", func(c jwt.MapClaims) { delete(c, "iat") }, jwt.SigningMethodHS256, "secret", domainmodelserror.TokenInvalid},
		{"no not before", func(c jwt.MapClaims) { delete(c, "nbf") }, jwt.SigningMethodHS256, "secret", domainmodelserror.TokenInvalid},
		{"subject mismatch", func(c jwt.MapClaims) { c["sub"] = "9" }, jwt.SigningMethodHS256, "secret", domainmodelserror.TokenInvalid},
		{"scope mismatch", func(c jwt.MapClaims) { c["space_id"] = 9 }, jwt.SigningMethodHS256, "secret", domainmodelserror.TokenInvalid},
		{"invalid membership", func(c jwt.MapClaims) { c["space_member_id"] = 0 }, jwt.SigningMethodHS256, "secret", domainmodelserror.TokenInvalid},
		{"future", func(c jwt.MapClaims) { c["nbf"] = time.Now().Add(time.Hour).Unix() }, jwt.SigningMethodHS256, "secret", domainmodelserror.TokenInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			now := time.Now().Unix()
			claims := jwt.MapClaims{"user_id": 1, "space_id": 2, "space_member_id": 3, "sub": "1", "aud": "lockmate:space:2", "token_type": "access", "iat": now, "nbf": now, "exp": now + 60}
			test.change(claims)
			signed, err := jwt.NewWithClaims(test.method, claims).SignedString([]byte(test.secret))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := token.ValidateAccess(2, signed); !errors.Is(err, test.expected) {
				t.Fatalf("expected %v got %v", test.expected, err)
			}
		})
	}
	if _, err := token.ValidateAccess(2, "malformed"); !errors.Is(err, domainmodelserror.TokenInvalid) || !errors.Is(err, jwt.ErrTokenMalformed) {
		t.Fatalf("lost malformed cause: %v", err)
	}
	if _, err := token.GenerateAccess(nil); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("nil claims accepted")
	}
	if _, err := token.GenerateRefresh(&domainmodelsentity.RefreshClaims{}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("invalid identity accepted")
	}
	for _, test := range []struct {
		secret   string
		duration time.Duration
	}{{"", time.Minute}, {"secret", 0}, {"secret", time.Millisecond}} {
		invalid := infrastructureutilitytoken.NewJwtImpl(test.secret, "refresh", test.duration, time.Hour)
		if _, err := invalid.GenerateAccess(&domainmodelsentity.AccessClaims{UserId: 1, SpaceId: 2, SpaceMemberId: 3}); !errors.Is(err, domainmodelserror.TokenFailed) {
			t.Fatalf("invalid configuration accepted: %v", err)
		}
	}
}
