package applicationseeder

import (
	"context"
	"errors"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
)

func (u *usecase) seedUsers(ctx context.Context, spaceId int64, roleIds map[string]int64, created *changes) error {
	const tag = tagPath + "/seedUsers"

	for _, value := range u.data.Users {
		auth, err := u.classicAuth.ReadByUsername(ctx, value.Username)
		var userId int64
		if err == nil {
			if auth == nil {
				err := domainmodelserror.NewError("seed credential read returned no record", domainmodelserror.Failure, nil)
				u.logger.Error(ctx, tag, "seed credential read returned no record", domainmodelslogger.Meta{"err": err})

				return err
			}

			userId = auth.UserId
		} else if errors.Is(err, domainmodelserror.ClassicAuthNotFound) {
			hash, err := u.password.Hash(value.Password)
			if err != nil {
				u.logger.Error(ctx, tag, "failed to hash seed user password", domainmodelslogger.Meta{"err": err})

				return err
			}

			userId, err = u.user.Create(ctx, domaincontractsrepository.CreateUserArgs{
				Name: value.Name,
				Bio:  value.Bio,
			})
			if err != nil {
				u.logger.Error(ctx, tag, "failed to create seed user profile", domainmodelslogger.Meta{"err": err})

				return err
			}

			created.users = append(created.users, record{id: userId, identifier: value.Username})
			authId, err := u.classicAuth.Create(ctx, domaincontractsrepository.CreateClassicAuthArgs{
				UserId:       userId,
				Username:     value.Username,
				PasswordHash: hash,
			})
			if err != nil {
				if canRetrySeed(ctx, err) {
					u.logger.Warn(ctx, tag, "failed to create seed classic credentials", domainmodelslogger.Meta{"err": err})
				} else {
					u.logger.Error(ctx, tag, "failed to create seed classic credentials", domainmodelslogger.Meta{"err": err})
				}

				return err
			}

			created.classicAuths = append(created.classicAuths, record{id: authId, identifier: value.Username})
		} else {
			u.logger.Error(ctx, tag, "failed to read seed classic credentials", domainmodelslogger.Meta{"err": err})

			return err
		}

		member, err := u.spaceMember.ReadByUserId(ctx, spaceId, userId)
		var memberId int64
		if err == nil {
			if member == nil {
				err := domainmodelserror.NewError("seed membership read returned no record", domainmodelserror.Failure, nil)
				u.logger.Error(ctx, tag, "seed membership read returned no record", domainmodelslogger.Meta{"err": err})

				return err
			}

			memberId = member.Id
		} else if errors.Is(err, domainmodelserror.SpaceMemberNotFound) {
			memberId, err = u.spaceMember.Create(ctx, spaceId, domaincontractsrepository.CreateSpaceMemberArgs{
				UserId: userId,
			})
			if err != nil {
				if canRetrySeed(ctx, err) {
					u.logger.Warn(ctx, tag, "failed to create seed space membership", domainmodelslogger.Meta{"err": err})
				} else {
					u.logger.Error(ctx, tag, "failed to create seed space membership", domainmodelslogger.Meta{"err": err})
				}

				return err
			}

			created.members = append(created.members, record{id: memberId, identifier: value.Username})
		} else {
			u.logger.Error(ctx, tag, "failed to read seed space membership", domainmodelslogger.Meta{"err": err})

			return err
		}

		roleId := roleIds[value.RoleSlug]
		_, err = u.spaceMemberRole.ReadBySpaceMemberIdAndRoleId(ctx, spaceId, memberId, roleId)
		if err == nil {
			continue
		}

		if !errors.Is(err, domainmodelserror.SpaceMemberRoleNotFound) {
			u.logger.Error(ctx, tag, "failed to read seed member-role assignment", domainmodelslogger.Meta{"err": err})

			return err
		}

		id, err := u.spaceMemberRole.Create(ctx, spaceId, domaincontractsrepository.CreateSpaceMemberRoleArgs{
			SpaceMemberId: memberId,
			RoleId:        roleId,
		})
		if err != nil {
			if canRetrySeed(ctx, err) {
				u.logger.Warn(ctx, tag, "failed to create seed member-role assignment", domainmodelslogger.Meta{"err": err})
			} else {
				u.logger.Error(ctx, tag, "failed to create seed member-role assignment", domainmodelslogger.Meta{"err": err})
			}

			return err
		}

		created.memberRoles = append(created.memberRoles, record{id: id, identifier: value.Username + " / " + value.RoleSlug})
	}

	return nil
}
