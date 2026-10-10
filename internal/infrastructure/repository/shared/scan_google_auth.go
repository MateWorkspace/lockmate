package infrastructurerepositoryshared

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	"github.com/jackc/pgx/v5"
)

func ScanGoogleAuth(row pgx.Row) (domainmodelsentity.GoogleAuth, error) {
	var item domainmodelsentity.GoogleAuth
	err := row.Scan(
		&item.Id,
		&item.UserId,
		&item.Subject,
		&item.Email,
		&item.IsEmailVerified,
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

func ScanGoogleAuthList(rows pgx.Rows) ([]domainmodelsentity.GoogleAuth, error) {
	items := make([]domainmodelsentity.GoogleAuth, 0)
	for rows.Next() {
		item, err := ScanGoogleAuth(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
