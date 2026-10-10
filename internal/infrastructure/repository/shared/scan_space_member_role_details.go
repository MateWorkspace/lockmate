package infrastructurerepositoryshared

import (
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	"github.com/jackc/pgx/v5"
)

func ScanSpaceMemberRoleDetails(row pgx.Row) (domaincontractsrepository.SpaceMemberRoleDetails, error) {
	var item domaincontractsrepository.SpaceMemberRoleDetails
	err := row.Scan(
		&item.SpaceMemberRole.Id,
		&item.SpaceMemberRole.SpaceId,
		&item.SpaceMemberRole.SpaceMemberId,
		&item.SpaceMemberRole.RoleId,
		&item.SpaceMemberRole.AuditCreate.At,
		&item.SpaceMemberRole.AuditCreate.By,
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
		&item.Role.Id,
		&item.Role.SpaceId,
		&item.Role.Slug,
		&item.Role.Name,
		&item.Role.Description,
		&item.Role.IsDefault,
		&item.Role.Preferences,
		&item.Role.AuditCreate.At,
		&item.Role.AuditUpdate.At,
		&item.Role.AuditDelete.At,
		&item.Role.AuditCreate.By,
		&item.Role.AuditUpdate.By,
		&item.Role.AuditDelete.By,
	)
	return item, err
}

func ScanSpaceMemberRoleDetailsList(rows pgx.Rows) ([]domaincontractsrepository.SpaceMemberRoleDetails, error) {
	items := make([]domaincontractsrepository.SpaceMemberRoleDetails, 0)
	for rows.Next() {
		item, err := ScanSpaceMemberRoleDetails(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func ScanSpaceMemberRoleWithSpaceMember(row pgx.Row) (domaincontractsrepository.SpaceMemberRoleWithSpaceMember, error) {
	var item domaincontractsrepository.SpaceMemberRoleWithSpaceMember
	err := row.Scan(
		&item.SpaceMemberRole.Id,
		&item.SpaceMemberRole.SpaceId,
		&item.SpaceMemberRole.SpaceMemberId,
		&item.SpaceMemberRole.RoleId,
		&item.SpaceMemberRole.AuditCreate.At,
		&item.SpaceMemberRole.AuditCreate.By,
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
	)
	return item, err
}

func ScanSpaceMemberRoleWithSpaceMemberList(rows pgx.Rows) ([]domaincontractsrepository.SpaceMemberRoleWithSpaceMember, error) {
	items := make([]domaincontractsrepository.SpaceMemberRoleWithSpaceMember, 0)
	for rows.Next() {
		item, err := ScanSpaceMemberRoleWithSpaceMember(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func ScanSpaceMemberRoleWithRole(row pgx.Row) (domaincontractsrepository.SpaceMemberRoleWithRole, error) {
	var item domaincontractsrepository.SpaceMemberRoleWithRole
	err := row.Scan(
		&item.SpaceMemberRole.Id,
		&item.SpaceMemberRole.SpaceId,
		&item.SpaceMemberRole.SpaceMemberId,
		&item.SpaceMemberRole.RoleId,
		&item.SpaceMemberRole.AuditCreate.At,
		&item.SpaceMemberRole.AuditCreate.By,
		&item.Role.Id,
		&item.Role.SpaceId,
		&item.Role.Slug,
		&item.Role.Name,
		&item.Role.Description,
		&item.Role.IsDefault,
		&item.Role.Preferences,
		&item.Role.AuditCreate.At,
		&item.Role.AuditUpdate.At,
		&item.Role.AuditDelete.At,
		&item.Role.AuditCreate.By,
		&item.Role.AuditUpdate.By,
		&item.Role.AuditDelete.By,
	)
	return item, err
}

func ScanSpaceMemberRoleWithRoleList(rows pgx.Rows) ([]domaincontractsrepository.SpaceMemberRoleWithRole, error) {
	items := make([]domaincontractsrepository.SpaceMemberRoleWithRole, 0)
	for rows.Next() {
		item, err := ScanSpaceMemberRoleWithRole(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
