package infrastructurerepositoryauthhandoff

import (
	"github.com/Masterminds/squirrel"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

var columns = []string{
	"code_hash",
	"client_id",
	"space_id",
	"user_id",
	"space_member_id",
	"redirect_uri",
	"is_new_user",
	"expires_at",
}

func (p *postgresImpl) queryCreate(value domainmodelsentity.AuthHandoff) (string, []any, error) {
	return p.sqr.Insert("auth_handoffs").Columns(columns...).Values(
		value.CodeHash,
		value.ClientId,
		value.SpaceId,
		value.UserId,
		value.SpaceMemberId,
		value.RedirectUri,
		value.IsNewUser,
		value.ExpiresAt,
	).ToSql()
}

func (p *postgresImpl) queryConsume(codeHash string, clientId int64) (string, []any, error) {
	q := p.sqr.Delete("auth_handoffs").Where(squirrel.Eq{"code_hash": codeHash, "client_id": clientId})
	return q.Suffix("RETURNING code_hash, client_id, space_id, user_id, space_member_id, redirect_uri, is_new_user, expires_at").ToSql()
}

func (p *postgresImpl) queryPurge(limit int64) (string, []any, error) {
	expired := p.sqr.Select("code_hash").From("auth_handoffs").Where("expires_at <= CURRENT_TIMESTAMP").OrderBy("expires_at").Limit(uint64(limit))
	return p.sqr.Delete("auth_handoffs").Where(squirrel.Expr("code_hash IN (?)", expired)).ToSql()
}
