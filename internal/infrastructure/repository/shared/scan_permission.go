package infrastructurerepositoryshared

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanPermission(row pgx.Row) (domainmodelsentity.Permission, error) {
	var item domainmodelsentity.Permission
	err := row.Scan(
		&item.Id,
		&item.SpaceId,
		&item.Slug,
		&item.Name,
		&item.Description,
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

func ScanPermissionList(rows pgx.Rows) ([]domainmodelsentity.Permission, error) {
	items := make([]domainmodelsentity.Permission, 0)
	for rows.Next() {
		item, err := ScanPermission(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
