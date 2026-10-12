package domainmodelsentity

import "encoding/json"

type Client struct {
	Id                int64           `db:"id" json:"id"`
	SpaceId           int64           `db:"space_id" json:"space_id"`
	Slug              string          `db:"slug" json:"slug"`
	Name              string          `db:"name" json:"name"`
	Description       string          `db:"description" json:"description"`
	SecretKeyHash     string          `db:"secret_key_hash" json:"secret_key_hash"`
	AuthRedirectUri   *string         `db:"auth_redirect_uri" json:"auth_redirect_uri"`
	SecretKeyRedacted string          `db:"secret_key_redacted" json:"secret_key_redacted"`
	IsActive          bool            `db:"is_active" json:"is_active"`
	Preferences       json.RawMessage `db:"preferences" json:"preferences"`

	AuditCreateUpdateDelete
}

// ClientIdentity is produced by client authentication, never by request mapping.
// Usecases recheck the live client and space before trusting this identity.
// User tokens remain scoped to SpaceId rather than an individual client.
type ClientIdentity struct {
	Id      int64  `json:"id"`
	SpaceId int64  `json:"space_id"`
	Slug    string `json:"slug"`
}
