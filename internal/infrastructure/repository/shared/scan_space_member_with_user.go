package infrastructurerepositoryshared

import (
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	"github.com/jackc/pgx/v5"
)

func ScanSpaceMemberWithUser(row pgx.Row) (domaincontractsrepository.SpaceMemberWithUser, error) {
	var item domaincontractsrepository.SpaceMemberWithUser
	err := row.Scan(
		&item.SpaceMember.Id,
		&item.SpaceMember.SpaceId,
		&item.SpaceMember.UserId,
		&item.SpaceMember.IsActive,
		&item.SpaceMember.Preferences,
		&item.SpaceMember.AuditCreate.At,
		&item.SpaceMember.AuditUpdate.At,
		&item.SpaceMember.AuditDelete.At,
		&item.SpaceMember.AuditCreate.By,
		&item.SpaceMember.AuditUpdate.By,
		&item.SpaceMember.AuditDelete.By,
		&item.User.Id,
		&item.User.Name,
		&item.User.Bio,
		&item.User.AvatarPath,
		&item.User.Preferences,
		&item.User.AuditCreate.At,
		&item.User.AuditUpdate.At,
		&item.User.AuditDelete.At,
		&item.User.AuditCreate.By,
		&item.User.AuditUpdate.By,
		&item.User.AuditDelete.By,
	)
	return item, err
}

func ScanSpaceMemberWithUserList(rows pgx.Rows) ([]domaincontractsrepository.SpaceMemberWithUser, error) {
	items := make([]domaincontractsrepository.SpaceMemberWithUser, 0)
	for rows.Next() {
		item, err := ScanSpaceMemberWithUser(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
