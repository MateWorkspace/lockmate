package infrastructurerepositoryshared

import (
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanClassicAuth(row pgx.Row) (domainmodelsentity.ClassicAuth, error) {
	var item domainmodelsentity.ClassicAuth
	err := row.Scan(
		&item.Id,
		&item.UserId,
		&item.Username,
		&item.Email,
		&item.Phone,
		&item.PasswordHash,
		&item.IsEmailVerified,
		&item.IsPhoneVerified,
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

func ScanClassicAuthSummary(row pgx.Row) (domaincontractsrepository.ClassicAuthSummary, error) {
	var item domaincontractsrepository.ClassicAuthSummary
	err := row.Scan(
		&item.Id,
		&item.UserId,
		&item.Username,
		&item.Email,
		&item.Phone,
		&item.IsEmailVerified,
		&item.IsPhoneVerified,
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

func ScanClassicAuthSummaryList(rows pgx.Rows) ([]domaincontractsrepository.ClassicAuthSummary, error) {
	items := make([]domaincontractsrepository.ClassicAuthSummary, 0)
	for rows.Next() {
		item, err := ScanClassicAuthSummary(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
