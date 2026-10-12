package infrastructurerepositorysigningkey

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructurerepositoryshared "github.com/MateWorkspace/lockmate/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(dt pgxdt.Pgxdt, sqrQuestion, sqrDollar *squirrel.StatementBuilderType) domaincontractsrepository.SigningKey {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) mapError(err error) error {
	return infrastructurerepositoryshared.MapPgxError(
		"signing key operation failed",
		err,
		infrastructurerepositoryshared.ConflictMatch{Constraint: "uq_signing_keys_kid", Type: domainmodelserror.SigningKeyConflict},
		infrastructurerepositoryshared.ConflictMatch{Constraint: "uq_signing_keys_active", Type: domainmodelserror.SigningKeyConflict},
	)
}

func (p *postgresImpl) Create(ctx context.Context, value domainmodelsentity.SigningKey) (id int64, err error) {
	if value.Kid == "" || value.Algorithm == "" || len(value.PrivateKeyCipher) == 0 || len(value.PrivateKeyNonce) == 0 || value.KeyFingerprint == "" {
		return 0, domainmodelserror.NewError("signing key material is incomplete", domainmodelserror.BadArgs, nil)
	}

	if len(value.PublicJWK) == 0 || !json.Valid(value.PublicJWK) {
		return 0, domainmodelserror.NewError("public JWK must be valid JSON", domainmodelserror.BadArgs, nil)
	}

	if err := infrastructurerepositoryshared.OptionalIds(value.AuditCreate.By); err != nil {
		return 0, err
	}

	query, values, err := p.queryCreate(value)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build signing key creation", err)
	}

	if err := p.Dt.QueryRow(ctx, query, values...).Scan(&id); err != nil {
		return 0, p.mapError(err)
	}

	return id, nil
}

func (p *postgresImpl) ReadActive(ctx context.Context) (*domainmodelsentity.SigningKey, error) {
	query, values, err := p.queryReadActive()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build signing key read", err)
	}

	item, err := infrastructurerepositoryshared.ScanSigningKey(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("active signing key not found", err, domainmodelserror.SigningKeyNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByKid(ctx context.Context, kid string) (*domainmodelsentity.SigningKey, error) {
	if kid == "" {
		return nil, domainmodelserror.NewError("kid is required", domainmodelserror.BadArgs, nil)
	}

	query, values, err := p.queryReadByKid(kid)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build signing key read", err)
	}

	item, err := infrastructurerepositoryshared.ScanSigningKey(p.Dt.QueryRow(ctx, query, values...))
	if err != nil {
		return nil, infrastructurerepositoryshared.ReadError("signing key not found", err, domainmodelserror.SigningKeyNotFound)
	}

	return &item, nil
}

func (p *postgresImpl) ListVerifiable(ctx context.Context) ([]domainmodelsentity.SigningKey, error) {
	query, values, err := p.queryListVerifiable()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build signing key list", err)
	}

	rows, err := p.Dt.Query(ctx, query, values...)
	if err != nil {
		return nil, p.mapError(err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanSigningKeyList(rows)
	if err != nil {
		return nil, p.mapError(err)
	}

	return items, nil
}

func (p *postgresImpl) ListStale(ctx context.Context, fingerprint string) ([]domainmodelsentity.SigningKey, error) {
	if fingerprint == "" {
		return nil, domainmodelserror.NewError("fingerprint is required", domainmodelserror.BadArgs, nil)
	}

	query, values, err := p.queryListStale(fingerprint)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build signing key list", err)
	}

	rows, err := p.Dt.Query(ctx, query, values...)
	if err != nil {
		return nil, p.mapError(err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanSigningKeyList(rows)
	if err != nil {
		return nil, p.mapError(err)
	}

	return items, nil
}

func (p *postgresImpl) DemoteActive(ctx context.Context, retiringAt time.Time, by *int64) (int64, error) {
	if err := infrastructurerepositoryshared.OptionalIds(by); err != nil {
		return 0, err
	}

	query, values, err := p.queryDemoteActive(retiringAt, by)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build signing key demotion", err)
	}

	var id int64
	if err := p.Dt.QueryRow(ctx, query, values...).Scan(&id); err != nil {
		return 0, infrastructurerepositoryshared.ReadError("active signing key not found", err, domainmodelserror.SigningKeyNotFound)
	}

	return id, nil
}

func (p *postgresImpl) UpdateSealed(ctx context.Context, id int64, fingerprint string, cipher []byte, nonce []byte) error {
	if err := infrastructurerepositoryshared.PositiveIds(id); err != nil {
		return err
	}

	if fingerprint == "" || len(cipher) == 0 || len(nonce) == 0 {
		return domainmodelserror.NewError("sealed signing key material is incomplete", domainmodelserror.BadArgs, nil)
	}

	query, values, err := p.queryUpdateSealed(id, fingerprint, cipher, nonce)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build signing key update", err)
	}

	result, err := p.Dt.Exec(ctx, query, values...)
	if err != nil {
		return p.mapError(err)
	}

	if result.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("signing key not found", domainmodelserror.SigningKeyNotFound)
	}

	return nil
}

func (p *postgresImpl) DeleteRetiringBefore(ctx context.Context, cutoff time.Time, by *int64) (int64, error) {
	if err := infrastructurerepositoryshared.OptionalIds(by); err != nil {
		return 0, err
	}

	query, values, err := p.queryDeleteRetiringBefore(cutoff, by)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build signing key cleanup", err)
	}

	result, err := p.Dt.Exec(ctx, query, values...)
	if err != nil {
		return 0, p.mapError(err)
	}

	return result.RowsAffected(), nil
}
