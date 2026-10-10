package applicationauthgoogle

import (
	"context"
	"errors"

	applicationshared "github.com/MateWorkspace/lockmate/internal/application/shared"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
)

const tagPath = "auth/google"

type usecase struct {
	client          domaincontractsrepository.Client
	space           domaincontractsrepository.Space
	user            domaincontractsrepository.User
	spaceMember     domaincontractsrepository.SpaceMember
	spaceMemberRole domaincontractsrepository.SpaceMemberRole
	token           domaincontractsutility.Token
	validator       domaincontractsutility.Validator
	transactor      domaincontractsutility.Transactor
	logger          domaincontractsutility.Logger
	googleAuth      domaincontractsrepository.GoogleAuth
	oauth           domaincontractsutility.GoogleOAuth
	identity        *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	client domaincontractsrepository.Client,
	space domaincontractsrepository.Space,
	user domaincontractsrepository.User,
	spaceMember domaincontractsrepository.SpaceMember,
	spaceMemberRole domaincontractsrepository.SpaceMemberRole,
	token domaincontractsutility.Token,
	validator domaincontractsutility.Validator,
	transactor domaincontractsutility.Transactor,
	logger domaincontractsutility.Logger,
	googleAuth domaincontractsrepository.GoogleAuth,
	oauth domaincontractsutility.GoogleOAuth,
) domainusecasesauth.Google {
	return &usecase{
		client:          client,
		space:           space,
		user:            user,
		spaceMember:     spaceMember,
		spaceMemberRole: spaceMemberRole,
		token:           token,
		validator:       validator,
		transactor:      transactor,
		logger:          logger,
		googleAuth:      googleAuth,
		oauth:           oauth,
		identity:        applicationshared.NewIdentityReader(user, space, spaceMember, ""),
	}
}

func (u *usecase) BeginSignIn(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.BeginGoogleSignInRequest) (result *domainusecasesauth.BeginGoogleSignInResult, err error) {
	const tag = tagPath + "/BeginSignIn"

	if request.Binding == "" {
		failure := domainmodelserror.NewError("browser binding is required", domainmodelserror.BadArgs, nil)
		u.logger.Error(ctx, tag, "browser binding is required", domainmodelslogger.Meta{
			"err": failure,
		})

		return nil, failure
	}

	if _, err := applicationshared.ReadClient(ctx, u.client, u.space, client); err != nil {
		u.logger.Error(ctx, tag, "failed to validate client identity", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	url, err := u.oauth.Begin(ctx, domaincontractsutility.BeginGoogleOAuthArgs{
		ClientId: client.Id,
		SpaceId:  client.SpaceId,
		Binding:  request.Binding,
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to begin google sign-in", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	return &domainusecasesauth.BeginGoogleSignInResult{AuthorizationUrl: url}, nil
}

func (u *usecase) SignInCallback(ctx context.Context, request domainusecasesauth.GoogleSignInCallbackRequest) (result *domainusecasesauth.SessionResult, err error) {
	const tag = tagPath + "/SignInCallback"

	if request.State == "" || request.Binding == "" || (request.Code == "" && request.ProviderError == "") || (request.Code != "" && request.ProviderError != "") {
		failure := domainmodelserror.NewError("invalid google callback", domainmodelserror.BadArgs, nil)
		u.logger.Error(ctx, tag, "invalid google callback", domainmodelslogger.Meta{
			"err": failure,
		})

		return nil, failure
	}

	completed, err := u.oauth.Complete(ctx, domaincontractsutility.CompleteGoogleOAuthArgs{
		Code:          request.Code,
		State:         request.State,
		Binding:       request.Binding,
		ProviderError: request.ProviderError,
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete google sign-in provider exchange", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	if completed == nil || completed.ClientId <= 0 || completed.SpaceId <= 0 || completed.Identity.Subject == "" || completed.Identity.Email == "" {
		failure := domainmodelserror.NewError("invalid google identity", domainmodelserror.GoogleOAuthInvalid, nil)
		u.logger.Error(ctx, tag, "invalid google identity", domainmodelslogger.Meta{
			"err": failure,
		})

		return nil, failure
	}

	changed := false
	for attempt := 0; attempt < 2; attempt++ {
		result, changed = nil, false
		err = u.transactor.WithTx(ctx, func(txCtx context.Context) error {
			client, err := u.client.ReadById(txCtx, completed.SpaceId, completed.ClientId)
			if err != nil {
				failure := applicationshared.IdentityError(err)
				u.logger.Error(txCtx, tag, "failed to read client by id", domainmodelslogger.Meta{
					"err": failure,
				})

				return failure
			}

			if client == nil {
				failure := applicationshared.Unauthorized(nil)
				u.logger.Error(txCtx, tag, "failed to read client by id", domainmodelslogger.Meta{
					"err": failure,
				})

				return failure
			}

			if _, err := applicationshared.ReadClient(txCtx, u.client, u.space, &domainmodelsentity.ClientIdentity{
				Id:      completed.ClientId,
				SpaceId: completed.SpaceId,
				Slug:    client.Slug,
			}); err != nil {
				u.logger.Error(txCtx, tag, "failed to validate client identity", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}

			auth, err := u.googleAuth.ReadBySubject(txCtx, completed.Identity.Subject)
			newUser := errors.Is(err, domainmodelserror.GoogleAuthNotFound)
			if err != nil && !newUser {
				u.logger.Error(txCtx, tag, "failed to read google auth by subject", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}

			var userId int64
			var by *int64
			if newUser {
				name := "Google user"
				if completed.Identity.Name != nil {
					if err := u.validator.UserName(*completed.Identity.Name); err != nil {
						u.logger.Warn(txCtx, tag, "using default name after provider name validation failed", domainmodelslogger.Meta{
							"err":       err,
							"client_id": completed.ClientId,
							"space_id":  completed.SpaceId,
						})
					} else {
						name = *completed.Identity.Name
					}
				}
				userId, err = u.user.Create(txCtx, domaincontractsrepository.CreateUserArgs{Name: name})
				if err != nil {
					u.logger.Error(txCtx, tag, "failed to create user", domainmodelslogger.Meta{
						"err": err,
					})

					return err
				}

				verified := completed.Identity.IsEmailVerified
				if _, err := u.googleAuth.Create(txCtx, domaincontractsrepository.CreateGoogleAuthArgs{
					UserId:          userId,
					Subject:         completed.Identity.Subject,
					Email:           completed.Identity.Email,
					IsEmailVerified: &verified,
				}); err != nil {
					if attempt < 1 && errors.Is(err, domainmodelserror.GoogleAuthSubjectConflict) {
						u.logger.Warn(txCtx, tag, "failed to create google auth", domainmodelslogger.Meta{
							"err": err,
						})
					} else {
						u.logger.Error(txCtx, tag, "failed to create google auth", domainmodelslogger.Meta{
							"err": err,
						})
					}

					return err
				}
				changed = true
			} else {
				if auth == nil || auth.Subject != completed.Identity.Subject || auth.UserId <= 0 || auth.AuditDelete.At != nil {
					failure := applicationshared.Unauthorized(nil)
					u.logger.Error(txCtx, tag, "invalid authentication identity", domainmodelslogger.Meta{
						"err": failure,
					})

					return failure
				}
				userId = auth.UserId
				by = &userId
				if auth.Email != completed.Identity.Email || auth.IsEmailVerified != completed.Identity.IsEmailVerified {
					email, verified := completed.Identity.Email, completed.Identity.IsEmailVerified
					if err := u.googleAuth.UpdateById(txCtx, auth.Id, domaincontractsrepository.UpdateGoogleAuthArgs{Email: &email, IsEmailVerified: &verified, By: by}); err != nil {
						u.logger.Error(txCtx, tag, "failed to update google auth", domainmodelslogger.Meta{
							"err": err,
						})

						return err
					}
					changed = true
				}
			}

			memberId, enrolled, err := applicationshared.EnsureMembership(txCtx, u.spaceMember, userId, completed.SpaceId, by)
			if err != nil {
				if attempt < 1 && errors.Is(err, domainmodelserror.SpaceMemberConflict) {
					u.logger.Warn(txCtx, tag, "failed to ensure space membership", domainmodelslogger.Meta{
						"err": err,
					})
				} else {
					u.logger.Error(txCtx, tag, "failed to ensure space membership", domainmodelslogger.Meta{
						"err": err,
					})
				}

				return err
			}
			changed = changed || enrolled
			result, err = applicationshared.ResolveSession(txCtx, u.identity, u.spaceMemberRole, userId, completed.SpaceId, memberId)
			if err != nil {
				u.logger.Error(txCtx, tag, "failed to resolve session identity and grants", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}

			result.IsNewUser = newUser
			accessToken, err := u.token.GenerateAccess(applicationshared.SessionClaims(result))
			if err != nil {
				u.logger.Error(txCtx, tag, "failed to generate access token", domainmodelslogger.Meta{
					"err":             err,
					"user_id":         result.User.Id,
					"space_id":        result.Space.Id,
					"space_member_id": result.SpaceMember.Id,
				})

				return err
			}

			refreshToken, err := u.token.GenerateRefresh(&domainmodelsentity.RefreshClaims{
				UserId:        result.User.Id,
				SpaceId:       result.Space.Id,
				SpaceMemberId: result.SpaceMember.Id,
				Name:          result.User.Name,
			})
			if err != nil {
				u.logger.Error(txCtx, tag, "failed to generate refresh token", domainmodelslogger.Meta{
					"err":             err,
					"user_id":         result.User.Id,
					"space_id":        result.Space.Id,
					"space_member_id": result.SpaceMember.Id,
				})

				return err
			}

			result.AccessToken, result.RefreshToken = accessToken, refreshToken

			return nil
		})
		if !errors.Is(err, domainmodelserror.GoogleAuthSubjectConflict) && !errors.Is(err, domainmodelserror.SpaceMemberConflict) {
			break
		}

		if attempt < 1 {
			u.logger.Warn(ctx, tag, "retrying transaction after conflict", domainmodelslogger.Meta{
				"err":          err,
				"next_attempt": attempt + 2,
			})
		}
	}

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete google sign-in transaction", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	if changed {
		u.logger.Info(ctx, tag, "updated google session identity", domainmodelslogger.Meta{
			"user_id":         result.User.Id,
			"space_id":        result.Space.Id,
			"space_member_id": result.SpaceMember.Id,
			"is_new_user":     result.IsNewUser,
		})
	}

	return result, nil
}
