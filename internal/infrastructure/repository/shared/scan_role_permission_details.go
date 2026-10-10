package infrastructurerepositoryshared

import (
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	"github.com/jackc/pgx/v5"
)

func ScanRolePermissionDetails(row pgx.Row) (domaincontractsrepository.RolePermissionDetails, error) {
	var item domaincontractsrepository.RolePermissionDetails
	err := row.Scan(
		&item.RolePermission.Id,
		&item.RolePermission.SpaceId,
		&item.RolePermission.RoleId,
		&item.RolePermission.PermissionId,
		&item.RolePermission.AuditCreate.At,
		&item.RolePermission.AuditCreate.By,
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
		&item.Permission.Id,
		&item.Permission.SpaceId,
		&item.Permission.Slug,
		&item.Permission.Name,
		&item.Permission.Description,
		&item.Permission.Preferences,
		&item.Permission.AuditCreate.At,
		&item.Permission.AuditUpdate.At,
		&item.Permission.AuditDelete.At,
		&item.Permission.AuditCreate.By,
		&item.Permission.AuditUpdate.By,
		&item.Permission.AuditDelete.By,
	)
	return item, err
}

func ScanRolePermissionDetailsList(rows pgx.Rows) ([]domaincontractsrepository.RolePermissionDetails, error) {
	items := make([]domaincontractsrepository.RolePermissionDetails, 0)
	for rows.Next() {
		item, err := ScanRolePermissionDetails(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func ScanRolePermissionWithRole(row pgx.Row) (domaincontractsrepository.RolePermissionWithRole, error) {
	var item domaincontractsrepository.RolePermissionWithRole
	err := row.Scan(
		&item.RolePermission.Id,
		&item.RolePermission.SpaceId,
		&item.RolePermission.RoleId,
		&item.RolePermission.PermissionId,
		&item.RolePermission.AuditCreate.At,
		&item.RolePermission.AuditCreate.By,
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

func ScanRolePermissionWithRoleList(rows pgx.Rows) ([]domaincontractsrepository.RolePermissionWithRole, error) {
	items := make([]domaincontractsrepository.RolePermissionWithRole, 0)
	for rows.Next() {
		item, err := ScanRolePermissionWithRole(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func ScanRolePermissionWithPermission(row pgx.Row) (domaincontractsrepository.RolePermissionWithPermission, error) {
	var item domaincontractsrepository.RolePermissionWithPermission
	err := row.Scan(
		&item.RolePermission.Id,
		&item.RolePermission.SpaceId,
		&item.RolePermission.RoleId,
		&item.RolePermission.PermissionId,
		&item.RolePermission.AuditCreate.At,
		&item.RolePermission.AuditCreate.By,
		&item.Permission.Id,
		&item.Permission.SpaceId,
		&item.Permission.Slug,
		&item.Permission.Name,
		&item.Permission.Description,
		&item.Permission.Preferences,
		&item.Permission.AuditCreate.At,
		&item.Permission.AuditUpdate.At,
		&item.Permission.AuditDelete.At,
		&item.Permission.AuditCreate.By,
		&item.Permission.AuditUpdate.By,
		&item.Permission.AuditDelete.By,
	)
	return item, err
}

func ScanRolePermissionWithPermissionList(rows pgx.Rows) ([]domaincontractsrepository.RolePermissionWithPermission, error) {
	items := make([]domaincontractsrepository.RolePermissionWithPermission, 0)
	for rows.Next() {
		item, err := ScanRolePermissionWithPermission(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
