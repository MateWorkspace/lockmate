package infrastructurerepositoryshared

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanClientOrigin(row pgx.Row) (domainmodelsentity.ClientOrigin, error) {
	var item domainmodelsentity.ClientOrigin
	err := row.Scan(
		&item.Id,
		&item.ClientId,
		&item.Origin,
		&item.IsActive,
		&item.AuditCreate.At,
		&item.AuditUpdate.At,
		&item.AuditDelete.At,
		&item.AuditCreate.By,
		&item.AuditUpdate.By,
		&item.AuditDelete.By,
	)

	return item, err
}

func ScanClientOriginList(rows pgx.Rows) ([]domainmodelsentity.ClientOrigin, error) {
	items := make([]domainmodelsentity.ClientOrigin, 0)
	for rows.Next() {
		item, err := ScanClientOrigin(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
