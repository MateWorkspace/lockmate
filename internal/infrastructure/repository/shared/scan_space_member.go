package infrastructurerepositoryshared

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanSpaceMember(row pgx.Row) (domainmodelsentity.SpaceMember, error) {
	var item domainmodelsentity.SpaceMember
	err := row.Scan(
		&item.Id,
		&item.SpaceId,
		&item.UserId,
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

func ScanSpaceMemberList(rows pgx.Rows) ([]domainmodelsentity.SpaceMember, error) {
	items := make([]domainmodelsentity.SpaceMember, 0)
	for rows.Next() {
		item, err := ScanSpaceMember(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
