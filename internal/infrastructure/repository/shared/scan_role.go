package infrastructurerepositoryshared

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanRole(row pgx.Row) (domainmodelsentity.Role, error) {
	var item domainmodelsentity.Role
	err := row.Scan(
		&item.Id,
		&item.SpaceId,
		&item.Slug,
		&item.Name,
		&item.Description,
		&item.IsDefault,
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

func ScanRoleList(rows pgx.Rows) ([]domainmodelsentity.Role, error) {
	items := make([]domainmodelsentity.Role, 0)
	for rows.Next() {
		item, err := ScanRole(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
