package utility_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitysecret "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/secret"
)

func TestSecretSealOpenRoundTripAndReencryption(t *testing.T) {
	current := bytes.Repeat([]byte{1}, 32)
	previous := bytes.Repeat([]byte{2}, 32)
	secret, err := infrastructureutilitysecret.NewSealImpl(current, previous)
	if err != nil {
		t.Fatal(err)
	}

	cipher, nonce, err := secret.Seal([]byte("private-material"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := secret.Open(cipher, nonce, secret.Fingerprint())
	if err != nil || string(got) != "private-material" {
		t.Fatalf("round trip: %v %q", err, got)
	}

	if _, err := secret.Open(cipher, nonce, "unknown"); !errors.Is(err, domainmodelserror.SecretKeyInvalid) {
		t.Fatalf("unknown fingerprint accepted: %v", err)
	}

	tampered := append([]byte(nil), cipher...)
	tampered[0] ^= 0xff
	if _, err := secret.Open(tampered, nonce, secret.Fingerprint()); !errors.Is(err, domainmodelserror.SecretDecryptFailed) {
		t.Fatalf("tampered ciphertext accepted: %v", err)
	}

	// A value sealed under the previous key opens via the previous fingerprint.
	// This models the KEK rotation re-encryption path.
	old, err := infrastructureutilitysecret.NewSealImpl(previous, nil)
	if err != nil {
		t.Fatal(err)
	}
	oldCipher, oldNonce, err := old.Seal([]byte("old-material"))
	if err != nil {
		t.Fatal(err)
	}
	if secret.PreviousFingerprint() != old.Fingerprint() {
		t.Fatal("previous fingerprint mismatch")
	}
	reopened, err := secret.Open(oldCipher, oldNonce, old.Fingerprint())
	if err != nil || string(reopened) != "old-material" {
		t.Fatalf("previous key decryption: %v %q", err, reopened)
	}
}

func TestSecretKeypairAndJWK(t *testing.T) {
	secret, err := infrastructureutilitysecret.NewSealImpl(bytes.Repeat([]byte{3}, 32), nil)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, publicJWK, kid, algorithm, err := secret.GenerateKeypair()
	if err != nil {
		t.Fatal(err)
	}
	if algorithm != "ES256" || kid == "" || len(privateKey) == 0 || len(publicJWK) == 0 {
		t.Fatalf("incomplete keypair: %q %q", kid, algorithm)
	}

	var jwk map[string]string
	if err := json.Unmarshal(publicJWK, &jwk); err != nil {
		t.Fatal(err)
	}
	if jwk["kty"] != "EC" || jwk["crv"] != "P-256" || jwk["alg"] != "ES256" || jwk["kid"] != kid || jwk["x"] == "" || jwk["y"] == "" {
		t.Fatalf("unexpected JWK: %v", jwk)
	}
}

func TestSecretRejectsInvalidKeys(t *testing.T) {
	if _, err := infrastructureutilitysecret.NewSealImpl(bytes.Repeat([]byte{1}, 16), nil); !errors.Is(err, domainmodelserror.SecretKeyInvalid) {
		t.Fatal("short current key accepted")
	}
	if _, err := infrastructureutilitysecret.NewSealImpl(bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 16)); !errors.Is(err, domainmodelserror.SecretKeyInvalid) {
		t.Fatal("short previous key accepted")
	}
}
