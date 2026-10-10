package infrastructurerepositoryshared

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanSpace(row pgx.Row) (domainmodelsentity.Space, error) {
	var item domainmodelsentity.Space
	err := row.Scan(
		&item.Id,
		&item.Slug,
		&item.Name,
		&item.Description,
		&item.IsActive,
		&item.Preferences,
		&item.AuditCreate.At,
		&item.AuditUpdate.At,
		&item.AuditDelete.At,
		&item.AuditCreate.By,
		&item.AuditUpdate.By,
		&item.AuditDelete.By,
	)
	return item, err
}

func ScanSpaceList(rows pgx.Rows) ([]domainmodelsentity.Space, error) {
	items := make([]domainmodelsentity.Space, 0)
	for rows.Next() {
		item, err := ScanSpace(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
