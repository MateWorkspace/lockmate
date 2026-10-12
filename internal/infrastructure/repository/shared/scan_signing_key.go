package infrastructurerepositoryshared

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanSigningKey(row pgx.Row) (domainmodelsentity.SigningKey, error) {
	var item domainmodelsentity.SigningKey
	err := row.Scan(
		&item.Id,
		&item.Kid,
		&item.Algorithm,
		&item.PrivateKeyCipher,
		&item.PrivateKeyNonce,
		&item.KeyFingerprint,
		&item.PublicJWK,
		&item.Status,
		&item.NotBefore,
		&item.RetiringAt,
		&item.AuditCreate.At,
		&item.AuditUpdate.At,
		&item.AuditDelete.At,
		&item.AuditCreate.By,
		&item.AuditUpdate.By,
		&item.AuditDelete.By,
	)
	return item, err
}

func ScanSigningKeyList(rows pgx.Rows) ([]domainmodelsentity.SigningKey, error) {
	items := make([]domainmodelsentity.SigningKey, 0)
	for rows.Next() {
		item, err := ScanSigningKey(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
