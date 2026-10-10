package infrastructurerepositoryshared

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanUser(row pgx.Row) (domainmodelsentity.User, error) {
	var item domainmodelsentity.User
	err := row.Scan(
		&item.Id,
		&item.Name,
		&item.Bio,
		&item.AvatarPath,
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

func ScanUserList(rows pgx.Rows) ([]domainmodelsentity.User, error) {
	items := make([]domainmodelsentity.User, 0)
	for rows.Next() {
		item, err := ScanUser(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
