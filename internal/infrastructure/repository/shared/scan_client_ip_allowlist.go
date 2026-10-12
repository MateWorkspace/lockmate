package infrastructurerepositoryshared

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanClientIpAllowlist(row pgx.Row) (domainmodelsentity.ClientIpAllowlist, error) {
	var item domainmodelsentity.ClientIpAllowlist
	err := row.Scan(
		&item.Id,
		&item.ClientId,
		&item.Cidr,
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

func ScanClientIpAllowlistList(rows pgx.Rows) ([]domainmodelsentity.ClientIpAllowlist, error) {
	items := make([]domainmodelsentity.ClientIpAllowlist, 0)
	for rows.Next() {
		item, err := ScanClientIpAllowlist(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
