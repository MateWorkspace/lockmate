package domainmodelsentity

import (
	"encoding/json"
	"time"
)

const (
	SigningKeyStatusActive   = "active"
	SigningKeyStatusRetiring = "retiring"
)

// SigningKey is a persisted asymmetric token-signing key. The private half is
// stored sealed (AEAD) and never serialized; only the public JWK is exposed.
type SigningKey struct {
	Id               int64           `db:"id" json:"id"`
	Kid              string          `db:"kid" json:"kid"`
	Algorithm        string          `db:"algorithm" json:"algorithm"`
	PrivateKeyCipher []byte          `db:"private_key_cipher" json:"-"`
	PrivateKeyNonce  []byte          `db:"private_key_nonce" json:"-"`
	KeyFingerprint   string          `db:"key_fingerprint" json:"-"`
	PublicJWK        json.RawMessage `db:"public_jwk" json:"public_jwk"`

	Status     string     `db:"status" json:"status"`
	NotBefore  time.Time  `db:"not_before" json:"not_before"`
	RetiringAt *time.Time `db:"retiring_at" json:"retiring_at"`

	AuditCreateUpdateDelete
}
