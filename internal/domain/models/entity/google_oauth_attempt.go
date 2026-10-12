package domainmodelsentity

import "time"

// GoogleOAuthAttempt is an internal expiring record; never serialize it as an API response.
type GoogleOAuthAttempt struct {
	StateHash   string    `db:"state_hash" json:"-"`
	BindingHash string    `db:"binding_hash" json:"-"`
	NonceHash   string    `db:"nonce_hash" json:"-"`
	Verifier    string    `db:"verifier" json:"-"`
	ClientId    int64     `db:"client_id" json:"-"`
	SpaceId     int64     `db:"space_id" json:"-"`
	RedirectUri string    `db:"redirect_uri" json:"-"`
	ClientState string    `db:"client_state" json:"-"`
	ExpiresAt   time.Time `db:"expires_at" json:"-"`
}
