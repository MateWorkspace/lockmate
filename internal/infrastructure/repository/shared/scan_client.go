package infrastructurerepositoryshared

import (
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanClient(row pgx.Row) (domainmodelsentity.Client, error) {
	var item domainmodelsentity.Client
	err := row.Scan(
		&item.Id,
		&item.SpaceId,
		&item.Slug,
		&item.Name,
		&item.Description,
		&item.SecretKeyHash,
		&item.SecretKeyRedacted,
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

func ScanClientSummary(row pgx.Row) (domaincontractsrepository.ClientSummary, error) {
	var item domaincontractsrepository.ClientSummary
	err := row.Scan(
		&item.Id,
		&item.SpaceId,
		&item.Slug,
		&item.Name,
		&item.Description,
		&item.SecretKeyRedacted,
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

func ScanClientSummaryList(rows pgx.Rows) ([]domaincontractsrepository.ClientSummary, error) {
	items := make([]domaincontractsrepository.ClientSummary, 0)
	for rows.Next() {
		item, err := ScanClientSummary(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
