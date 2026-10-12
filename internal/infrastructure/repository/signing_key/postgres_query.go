package infrastructurerepositorysigningkey

import (
	"time"

	"github.com/Masterminds/squirrel"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
)

var columns = []string{
	"id",
	"kid",
	"algorithm",
	"private_key_cipher",
	"private_key_nonce",
	"key_fingerprint",
	"public_jwk",
	"status",
	"not_before",
	"retiring_at",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) querySelect() squirrel.SelectBuilder {
	return p.SqrD.Select(columns...).From("signing_keys").Where("deleted_at IS NULL")
}

func (p *postgresImpl) queryCreate(value domainmodelsentity.SigningKey) (string, []any, error) {
	return p.SqrD.Insert("signing_keys").Columns(
		"kid",
		"algorithm",
		"private_key_cipher",
		"private_key_nonce",
		"key_fingerprint",
		"public_jwk",
		"status",
		"not_before",
		"created_by",
	).Values(
		value.Kid,
		value.Algorithm,
		value.PrivateKeyCipher,
		value.PrivateKeyNonce,
		value.KeyFingerprint,
		value.PublicJWK,
		value.Status,
		value.NotBefore,
		value.AuditCreate.By,
	).Suffix("RETURNING id").ToSql()
}

func (p *postgresImpl) queryReadActive() (string, []any, error) {
	return p.querySelect().
		Where(squirrel.Eq{"status": domainmodelsentity.SigningKeyStatusActive}).
		OrderBy("id DESC").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadByKid(kid string) (string, []any, error) {
	return p.querySelect().Where(squirrel.Eq{"kid": kid}).ToSql()
}

func (p *postgresImpl) queryListVerifiable() (string, []any, error) {
	return p.querySelect().
		Where(squirrel.Eq{"status": []string{domainmodelsentity.SigningKeyStatusActive, domainmodelsentity.SigningKeyStatusRetiring}}).
		OrderBy("id ASC").
		ToSql()
}

func (p *postgresImpl) queryListStale(fingerprint string) (string, []any, error) {
	return p.querySelect().
		Where(squirrel.NotEq{"key_fingerprint": fingerprint}).
		OrderBy("id ASC").
		ToSql()
}

func (p *postgresImpl) queryDemoteActive(retiringAt time.Time, by *int64) (string, []any, error) {
	return p.SqrD.Update("signing_keys").
		Set("status", domainmodelsentity.SigningKeyStatusRetiring).
		Set("retiring_at", retiringAt).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", by).
		Where(squirrel.Eq{"status": domainmodelsentity.SigningKeyStatusActive}).
		Where("deleted_at IS NULL").
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryUpdateSealed(id int64, fingerprint string, cipher []byte, nonce []byte) (string, []any, error) {
	return p.SqrD.Update("signing_keys").
		Set("private_key_cipher", cipher).
		Set("private_key_nonce", nonce).
		Set("key_fingerprint", fingerprint).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryDeleteRetiringBefore(cutoff time.Time, by *int64) (string, []any, error) {
	return p.SqrD.Update("signing_keys").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", by).
		Where(squirrel.Eq{"status": domainmodelsentity.SigningKeyStatusRetiring}).
		Where("retiring_at IS NOT NULL").
		Where(squirrel.LtOrEq{"retiring_at": cutoff}).
		Where("deleted_at IS NULL").
		ToSql()
}
