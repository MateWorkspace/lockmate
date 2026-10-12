package utility_test

import (
	"crypto/ecdsa"
	"crypto/x509"
	"errors"
	"reflect"
	"testing"
	"time"

	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitysecret "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/secret"
	infrastructureutilitytoken "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/token"
	"github.com/golang-jwt/jwt/v5"
)

func newAccessKey(t *testing.T) (domainmodelsentity.SigningKey, []byte) {
	t.Helper()
	secret, err := infrastructureutilitysecret.NewSealImpl([]byte("0123456789abcdef0123456789abcdef"), nil)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, publicJWK, kid, algorithm, err := secret.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	return domainmodelsentity.SigningKey{Id: 1, Kid: kid, Algorithm: algorithm, PublicJWK: publicJWK, Status: domainmodelsentity.SigningKeyStatusActive}, privateKey
}

func TestTokenRoundTripAndPurpose(t *testing.T) {
	key, privateKey := newAccessKey(t)
	token := infrastructureutilitytoken.NewJwtImpl("refresh-secret", time.Minute, time.Hour)
	want := &domainmodelsentity.AccessClaims{UserId: 1, SpaceId: 2, SpaceMemberId: 3, Name: "Google User", SpaceName: "Lockmate", Roles: []string{"user"}, Permissions: []string{"profile:get"}}
	access, err := token.GenerateAccess(want, key.Kid, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	got, err := token.ValidateAccess([]domainmodelsentity.SigningKey{key}, access)
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
	if _, err := token.ValidateAccess([]domainmodelsentity.SigningKey{key}, refresh); !errors.Is(err, domainmodelserror.TokenInvalid) {
		t.Fatal("refresh used as access")
	}
	if _, err := token.ValidateRefresh(2, access); !errors.Is(err, domainmodelserror.TokenInvalid) {
		t.Fatal("access used as refresh")
	}

	other, _ := newAccessKey(t)
	if _, err := token.ValidateAccess([]domainmodelsentity.SigningKey{other}, access); !errors.Is(err, domainmodelserror.TokenInvalid) {
		t.Fatal("unknown signing key accepted")
	}
	if _, err := token.ValidateAccess([]domainmodelsentity.SigningKey{key}, access+"x"); !errors.Is(err, domainmodelserror.TokenInvalid) {
		t.Fatal("tampered token accepted")
	}
	if _, err := token.ValidateAccess([]domainmodelsentity.SigningKey{key}, "malformed"); !errors.Is(err, domainmodelserror.TokenInvalid) {
		t.Fatal("malformed token accepted")
	}
}

func TestTokenSigningAndAudienceFailures(t *testing.T) {
	key, privateKey := newAccessKey(t)
	token := infrastructureutilitytoken.NewJwtImpl("refresh-secret", time.Minute, time.Hour)

	if _, err := token.GenerateAccess(nil, key.Kid, privateKey); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("nil claims accepted")
	}
	if _, err := token.GenerateAccess(&domainmodelsentity.AccessClaims{UserId: 1, SpaceId: 2, SpaceMemberId: 3}, "", nil); !errors.Is(err, domainmodelserror.TokenFailed) {
		t.Fatal("missing signing key accepted")
	}
	if _, err := token.GenerateRefresh(&domainmodelsentity.RefreshClaims{}); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("invalid identity accepted")
	}
	invalid := infrastructureutilitytoken.NewJwtImpl("", time.Minute, time.Hour)
	if _, err := invalid.GenerateRefresh(&domainmodelsentity.RefreshClaims{UserId: 1, SpaceId: 2, SpaceMemberId: 3}); !errors.Is(err, domainmodelserror.TokenFailed) {
		t.Fatal("missing refresh secret accepted")
	}
	if _, err := token.ValidateAccess(nil, "token"); !errors.Is(err, domainmodelserror.BadArgs) {
		t.Fatal("missing verification keys accepted")
	}

	// A token whose audience does not match its own space claim is rejected.
	now := time.Now().Unix()
	claims := jwt.MapClaims{"user_id": 1, "space_id": 2, "space_member_id": 3, "sub": "1", "aud": "lockmate:space:9", "token_type": "access", "iat": now, "nbf": now, "exp": now + 60}
	signed, err := signES256(t, privateKey, key.Kid, claims)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := token.ValidateAccess([]domainmodelsentity.SigningKey{key}, signed); !errors.Is(err, domainmodelserror.TokenInvalid) {
		t.Fatalf("audience mismatch accepted: %v", err)
	}

	expired := jwt.MapClaims{"user_id": 1, "space_id": 2, "space_member_id": 3, "sub": "1", "aud": "lockmate:space:2", "token_type": "access", "iat": now - 7200, "nbf": now - 7200, "exp": now - 3600}
	signed, err = signES256(t, privateKey, key.Kid, expired)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := token.ValidateAccess([]domainmodelsentity.SigningKey{key}, signed); !errors.Is(err, domainmodelserror.TokenExpired) {
		t.Fatalf("expired token accepted: %v", err)
	}
}

func signES256(t *testing.T, privateKeyPKCS8 []byte, kid string, claims jwt.MapClaims) (string, error) {
	t.Helper()
	parsed, err := x509.ParsePKCS8PrivateKey(privateKeyPKCS8)
	if err != nil {
		t.Fatal(err)
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		t.Fatal("not an ECDSA key")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = kid
	return token.SignedString(key)
}
