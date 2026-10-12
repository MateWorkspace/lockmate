package applicationauthsession

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

const tagPath = "auth/session"

type usecase struct {
	client          domaincontractsrepository.Client
	space           domaincontractsrepository.Space
	user            domaincontractsrepository.User
	classicAuth     domaincontractsrepository.ClassicAuth
	spaceMember     domaincontractsrepository.SpaceMember
	spaceMemberRole domaincontractsrepository.SpaceMemberRole
	signingKeys     domaincontractsrepository.SigningKey
	password        domaincontractsutility.Password
	token           domaincontractsutility.Token
	secret          domaincontractsutility.Secret
	validator       domaincontractsutility.Validator
	transactor      domaincontractsutility.Transactor
	logger          domaincontractsutility.Logger
	identity        *applicationshared.IdentityReader
}

func NewUsecaseImpl(
	client domaincontractsrepository.Client,
	space domaincontractsrepository.Space,
	user domaincontractsrepository.User,
	classicAuth domaincontractsrepository.ClassicAuth,
	spaceMember domaincontractsrepository.SpaceMember,
	spaceMemberRole domaincontractsrepository.SpaceMemberRole,
	signingKeys domaincontractsrepository.SigningKey,
	password domaincontractsutility.Password,
	token domaincontractsutility.Token,
	secret domaincontractsutility.Secret,
	validator domaincontractsutility.Validator,
	transactor domaincontractsutility.Transactor,
	logger domaincontractsutility.Logger,
) domainusecasesauth.Session {
	return &usecase{
		client:          client,
		space:           space,
		user:            user,
		classicAuth:     classicAuth,
		spaceMember:     spaceMember,
		spaceMemberRole: spaceMemberRole,
		signingKeys:     signingKeys,
		password:        password,
		token:           token,
		secret:          secret,
		validator:       validator,
		transactor:      transactor,
		logger:          logger,
		identity:        applicationshared.NewIdentityReader(user, space, spaceMember, ""),
	}
}

func (u *usecase) SignIn(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.SignInRequest) (result *domainusecasesauth.TokenResult, err error) {
	const tag = tagPath + "/SignIn"

	if request.Username == "" || request.Password == "" {
		failure := domainmodelserror.NewError("credentials are required", domainmodelserror.BadArgs, nil)
		u.logger.Error(ctx, tag, "credentials are required", domainmodelslogger.Meta{
			"err": failure,
		})

		return nil, failure
	}

	var session *domainusecasesauth.SessionResult
	enrolled := false
	for attempt := 0; attempt < 2; attempt++ {
		session, result, enrolled = nil, nil, false
		err = u.transactor.WithTx(ctx, func(txCtx context.Context) error {
			if _, err := applicationshared.ReadClient(txCtx, u.client, u.space, client); err != nil {
				u.logger.Error(txCtx, tag, "failed to validate client identity", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}

			auth, err := u.classicAuth.ReadByUsername(txCtx, request.Username)
			if errors.Is(err, domainmodelserror.ClassicAuthNotFound) {
				failure := applicationshared.Unauthorized(err)
				u.logger.Error(txCtx, tag, "failed to read classic auth by username", domainmodelslogger.Meta{
					"err": failure,
				})

				return failure
			}

			if err != nil {
				u.logger.Error(txCtx, tag, "failed to read classic auth by username", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}

			if auth == nil || auth.Username != request.Username || auth.UserId <= 0 || auth.AuditDelete.At != nil {
				failure := applicationshared.Unauthorized(nil)
				u.logger.Error(txCtx, tag, "invalid authentication identity", domainmodelslogger.Meta{
					"err": failure,
				})

				return failure
			}

			if err := u.password.Compare(auth.PasswordHash, request.Password); err != nil {
				if errors.Is(err, domainmodelserror.PasswordMismatch) {
					failure := applicationshared.Unauthorized(err)
					u.logger.Error(txCtx, tag, "failed to compare password", domainmodelslogger.Meta{
						"err": failure,
					})

					return failure
				}

				u.logger.Error(txCtx, tag, "failed to compare password", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}

			memberId, created, err := applicationshared.EnsureMembership(txCtx, u.spaceMember, auth.UserId, client.SpaceId, &auth.UserId)
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
			enrolled = created
			session, err = applicationshared.ResolveSession(txCtx, u.identity, u.spaceMemberRole, auth.UserId, client.SpaceId, memberId)
			if err != nil {
				u.logger.Error(txCtx, tag, "failed to resolve session identity and grants", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}

			kid, privateKey, err := applicationshared.ResolveSigningKey(txCtx, u.signingKeys, u.secret)
			if err != nil {
				u.logger.Error(txCtx, tag, "failed to resolve access signing key", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}

			accessToken, err := u.token.GenerateAccess(applicationshared.SessionClaims(session), kid, privateKey)
			if err != nil {
				u.logger.Error(txCtx, tag, "failed to generate access token", domainmodelslogger.Meta{
					"err":             err,
					"user_id":         session.User.Id,
					"space_id":        session.Space.Id,
					"space_member_id": session.SpaceMember.Id,
				})

				return err
			}

			refreshToken, err := u.token.GenerateRefresh(&domainmodelsentity.RefreshClaims{
				UserId:        session.User.Id,
				SpaceId:       session.Space.Id,
				SpaceMemberId: session.SpaceMember.Id,
				Name:          session.User.Name,
			})
			if err != nil {
				u.logger.Error(txCtx, tag, "failed to generate refresh token", domainmodelslogger.Meta{
					"err":             err,
					"user_id":         session.User.Id,
					"space_id":        session.Space.Id,
					"space_member_id": session.SpaceMember.Id,
				})

				return err
			}

			result = &domainusecasesauth.TokenResult{AccessToken: accessToken, RefreshToken: refreshToken, IsNewUser: session.IsNewUser}

			return nil
		})
		if !errors.Is(err, domainmodelserror.SpaceMemberConflict) {
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
		u.logger.Error(ctx, tag, "failed to complete sign-in transaction", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	if enrolled {
		u.logger.Info(ctx, tag, "created membership", domainmodelslogger.Meta{
			"user_id":         session.User.Id,
			"space_id":        session.Space.Id,
			"space_member_id": session.SpaceMember.Id,
		})
	}

	return result, nil
}

func (u *usecase) Register(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.RegisterRequest) (result *domainusecasesauth.TokenResult, err error) {
	const tag = tagPath + "/Register"

	var session *domainusecasesauth.SessionResult
	err = u.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := applicationshared.ReadClient(txCtx, u.client, u.space, client); err != nil {
			u.logger.Error(txCtx, tag, "failed to validate client identity", domainmodelslogger.Meta{
				"err": err,
			})

			return err
		}

		if err := u.validator.UserName(request.Name); err != nil {
			u.logger.Error(txCtx, tag, "invalid user name", domainmodelslogger.Meta{
				"err": err,
			})

			return err
		}

		if err := u.validator.UserUsername(request.Username); err != nil {
			u.logger.Error(txCtx, tag, "invalid user username", domainmodelslogger.Meta{
				"err": err,
			})

			return err
		}

		if err := u.validator.UserPassword(request.Password); err != nil {
			u.logger.Error(txCtx, tag, "invalid user password", domainmodelslogger.Meta{
				"err": err,
			})

			return err
		}

		if request.Bio != nil {
			if err := u.validator.UserBio(*request.Bio); err != nil {
				u.logger.Error(txCtx, tag, "invalid user bio", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}
		}

		if request.Email != nil {
			if err := u.validator.UserEmail(*request.Email); err != nil {
				u.logger.Error(txCtx, tag, "invalid user email", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}
		}

		if request.Phone != nil {
			if err := u.validator.UserPhone(*request.Phone); err != nil {
				u.logger.Error(txCtx, tag, "invalid user phone", domainmodelslogger.Meta{
					"err": err,
				})

				return err
			}
		}

		hash, err := u.password.Hash(request.Password)
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to hash password", domainmodelslogger.Meta{
				"err": err,
			})

			return err
		}

		userId, err := u.user.Create(txCtx, domaincontractsrepository.CreateUserArgs{Name: request.Name, Bio: request.Bio})
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to create user", domainmodelslogger.Meta{
				"err": err,
			})

			return err
		}

		if _, err := u.classicAuth.Create(txCtx, domaincontractsrepository.CreateClassicAuthArgs{
			UserId:       userId,
			Username:     request.Username,
			PasswordHash: hash,
			Email:        request.Email,
			Phone:        request.Phone,
		}); err != nil {
			u.logger.Error(txCtx, tag, "failed to create classic auth", domainmodelslogger.Meta{
				"err": err,
			})

			return err
		}

		memberId, err := u.spaceMember.Create(txCtx, client.SpaceId, domaincontractsrepository.CreateSpaceMemberArgs{UserId: userId})
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to create space member", domainmodelslogger.Meta{
				"err": err,
			})

			return err
		}

		session, err = applicationshared.ResolveSession(txCtx, u.identity, u.spaceMemberRole, userId, client.SpaceId, memberId)
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to resolve session identity and grants", domainmodelslogger.Meta{
				"err": err,
			})

			return err
		}

		session.IsNewUser = true

		kid, privateKey, err := applicationshared.ResolveSigningKey(txCtx, u.signingKeys, u.secret)
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to resolve access signing key", domainmodelslogger.Meta{
				"err": err,
			})

			return err
		}

		accessToken, err := u.token.GenerateAccess(applicationshared.SessionClaims(session), kid, privateKey)
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to generate access token", domainmodelslogger.Meta{
				"err":             err,
				"user_id":         session.User.Id,
				"space_id":        session.Space.Id,
				"space_member_id": session.SpaceMember.Id,
			})

			return err
		}

		refreshToken, err := u.token.GenerateRefresh(&domainmodelsentity.RefreshClaims{
			UserId:        session.User.Id,
			SpaceId:       session.Space.Id,
			SpaceMemberId: session.SpaceMember.Id,
			Name:          session.User.Name,
		})
		if err != nil {
			u.logger.Error(txCtx, tag, "failed to generate refresh token", domainmodelslogger.Meta{
				"err":             err,
				"user_id":         session.User.Id,
				"space_id":        session.Space.Id,
				"space_member_id": session.SpaceMember.Id,
			})

			return err
		}

		result = &domainusecasesauth.TokenResult{AccessToken: accessToken, RefreshToken: refreshToken, IsNewUser: session.IsNewUser}

		return nil
	})

	if err != nil {
		u.logger.Error(ctx, tag, "failed to complete registration transaction", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	u.logger.Info(ctx, tag, "registered user", domainmodelslogger.Meta{
		"user_id":         session.User.Id,
		"space_id":        session.Space.Id,
		"space_member_id": session.SpaceMember.Id,
	})
	return result, nil
}

func (u *usecase) Refresh(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.RefreshRequest) (result *domainusecasesauth.TokenResult, err error) {
	const tag = tagPath + "/Refresh"

	if request.RefreshToken == "" {
		failure := domainmodelserror.NewError("refresh token is required", domainmodelserror.BadArgs, nil)
		u.logger.Error(ctx, tag, "refresh token is required", domainmodelslogger.Meta{
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

	claims, err := u.token.ValidateRefresh(client.SpaceId, request.RefreshToken)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to validate refresh token", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	if claims == nil || claims.SpaceId != client.SpaceId {
		failure := applicationshared.Unauthorized(nil)
		u.logger.Error(ctx, tag, "refresh token space does not match client", domainmodelslogger.Meta{
			"err": failure,
		})

		return nil, failure
	}

	session, err := applicationshared.ResolveSession(ctx, u.identity, u.spaceMemberRole, claims.UserId, claims.SpaceId, claims.SpaceMemberId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve session identity and grants", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	kid, privateKey, err := applicationshared.ResolveSigningKey(ctx, u.signingKeys, u.secret)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve access signing key", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	accessToken, err := u.token.GenerateAccess(applicationshared.SessionClaims(session), kid, privateKey)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to generate access token", domainmodelslogger.Meta{
			"err":             err,
			"user_id":         session.User.Id,
			"space_id":        session.Space.Id,
			"space_member_id": session.SpaceMember.Id,
		})

		return nil, err
	}

	refreshToken, err := u.token.GenerateRefresh(&domainmodelsentity.RefreshClaims{
		UserId:        session.User.Id,
		SpaceId:       session.Space.Id,
		SpaceMemberId: session.SpaceMember.Id,
		Name:          session.User.Name,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to generate refresh token", domainmodelslogger.Meta{
			"err":             err,
			"user_id":         session.User.Id,
			"space_id":        session.Space.Id,
			"space_member_id": session.SpaceMember.Id,
		})

		return nil, err
	}

	return &domainusecasesauth.TokenResult{AccessToken: accessToken, RefreshToken: refreshToken, IsNewUser: false}, nil
}

func (u *usecase) AuthenticateAccessToken(ctx context.Context, client *domainmodelsentity.ClientIdentity, accessToken string) (result *domainmodelsentity.AccessClaims, err error) {
	const tag = tagPath + "/AuthenticateAccessToken"

	if accessToken == "" {
		failure := domainmodelserror.NewError("access token is required", domainmodelserror.BadArgs, nil)
		u.logger.Error(ctx, tag, "access token is required", domainmodelslogger.Meta{
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

	keys, err := applicationshared.VerifiableKeys(ctx, u.signingKeys)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read verifiable signing keys", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	claims, err := u.token.ValidateAccess(keys, accessToken)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to validate access token", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	if claims == nil || claims.SpaceId != client.SpaceId {
		failure := applicationshared.Unauthorized(nil)
		u.logger.Error(ctx, tag, "access token space does not match client", domainmodelslogger.Meta{
			"err": failure,
		})

		return nil, failure
	}

	session, err := applicationshared.ResolveSession(ctx, u.identity, u.spaceMemberRole, claims.UserId, claims.SpaceId, claims.SpaceMemberId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve session identity and grants", domainmodelslogger.Meta{
			"err": err,
		})

		return nil, err
	}

	return applicationshared.SessionClaims(session), nil
}
