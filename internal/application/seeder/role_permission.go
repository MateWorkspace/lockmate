package applicationseeder

import (
	"context"
	"errors"
	"fmt"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

func (u *usecase) seedRolePermissions(ctx context.Context, spaceId int64, roleIds, permissionIds map[string]int64, created *changes) error {
	const tag = tagPath + "/seedRolePermissions"

	for _, role := range u.data.Roles {
		roleId := roleIds[role.Slug]
		for _, slug := range role.Permissions {
			permissionId := permissionIds[slug]
			_, err := u.rolePermission.ReadByRoleIdAndPermissionId(ctx, spaceId, roleId, permissionId)
			if err == nil {
				continue
			}

			if !errors.Is(err, domainmodelserror.RolePermissionNotFound) {
				u.logger.Error(ctx, tag, "failed to read seed role-permission assignment", domainmodelslogger.Meta{"err": err})

				return err
			}

			id, err := u.rolePermission.Create(ctx, spaceId, domaincontractsrepository.CreateRolePermissionArgs{
				RoleId:       roleId,
				PermissionId: permissionId,
			})
			if err != nil {
				if canRetrySeed(ctx, err) {
					u.logger.Warn(ctx, tag, "failed to create seed role-permission assignment", domainmodelslogger.Meta{"err": err})
				} else {
					u.logger.Error(ctx, tag, "failed to create seed role-permission assignment", domainmodelslogger.Meta{"err": err})
				}

				return err
			}

			created.rolePermissions = append(created.rolePermissions, record{id: id, identifier: fmt.Sprintf("%s / %s", role.Slug, slug)})
		}
	}

	return nil
}
