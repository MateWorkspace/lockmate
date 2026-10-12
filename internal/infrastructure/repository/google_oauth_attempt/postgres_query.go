package infrastructurerepositorygoogleoauthattempt

import (
	"github.com/Masterminds/squirrel"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

var columns = []string{
	"state_hash",
	"binding_hash",
	"nonce_hash",
	"verifier",
	"client_id",
	"space_id",
	"redirect_uri",
	"client_state",
	"expires_at",
}

func (p *postgresImpl) queryCreate(value domainmodelsentity.GoogleOAuthAttempt) (string, []any, error) {
	return p.sqr.Insert("google_oauth_attempts").Columns(columns...).Values(
		value.StateHash,
		value.BindingHash,
		value.NonceHash,
		value.Verifier,
		value.ClientId,
		value.SpaceId,
		value.RedirectUri,
		value.ClientState,
		value.ExpiresAt,
	).ToSql()
}

func (p *postgresImpl) queryConsume(stateHash string, bindingHash string) (string, []any, error) {
	q := p.sqr.Delete("google_oauth_attempts").Where(squirrel.Eq{"state_hash": stateHash, "binding_hash": bindingHash})
	q = q.Where("expires_at > CURRENT_TIMESTAMP")
	return q.Suffix("RETURNING state_hash, binding_hash, nonce_hash, verifier, client_id, space_id, redirect_uri, client_state, expires_at").ToSql()
}

func (p *postgresImpl) queryPurge(limit int64) (string, []any, error) {
	expired := p.sqr.Select("state_hash").From("google_oauth_attempts").Where("expires_at <= CURRENT_TIMESTAMP").OrderBy("expires_at").Limit(uint64(limit))
	return p.sqr.Delete("google_oauth_attempts").Where(squirrel.Expr("state_hash IN (?)", expired)).ToSql()
}
