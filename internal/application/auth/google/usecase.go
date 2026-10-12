package applicationauthgoogle

import (
	"context"
	"errors"
	"net/url"
	"time"

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
	validator       domaincontractsutility.Validator
	transactor      domaincontractsutility.Transactor
	logger          domaincontractsutility.Logger
	googleAuth      domaincontractsrepository.GoogleAuth
	oauth           domaincontractsutility.GoogleOAuth
	launches        domaincontractsrepository.GoogleOAuthLaunch
	handoffs        domaincontractsrepository.AuthHandoff
	flow            domaincontractsutility.AuthFlow
	config          Config
	identity        *applicationshared.IdentityReader
}

type Config struct {
	Enabled         bool
	PublicUrl       string
	LaunchDuration  time.Duration
	AttemptDuration time.Duration
	HandoffDuration time.Duration
}

func NewUsecaseImpl(
	client domaincontractsrepository.Client,
	space domaincontractsrepository.Space,
	user domaincontractsrepository.User,
	spaceMember domaincontractsrepository.SpaceMember,
	spaceMemberRole domaincontractsrepository.SpaceMemberRole,
	validator domaincontractsutility.Validator,
	transactor domaincontractsutility.Transactor,
	logger domaincontractsutility.Logger,
	googleAuth domaincontractsrepository.GoogleAuth,
	oauth domaincontractsutility.GoogleOAuth,
	launches domaincontractsrepository.GoogleOAuthLaunch,
	handoffs domaincontractsrepository.AuthHandoff,
	flow domaincontractsutility.AuthFlow,
	config Config,
) domainusecasesauth.Google {
	return &usecase{
		client:          client,
		space:           space,
		user:            user,
		spaceMember:     spaceMember,
		spaceMemberRole: spaceMemberRole,
		validator:       validator,
		transactor:      transactor,
		logger:          logger,
		googleAuth:      googleAuth,
		oauth:           oauth,
		launches:        launches,
		handoffs:        handoffs,
		flow:            flow,
		config:          config,
		identity:        applicationshared.NewIdentityReader(user, space, spaceMember, ""),
	}
}

func (u *usecase) SignInCallback(ctx context.Context, request domainusecasesauth.GoogleSignInCallbackRequest) (output *domainusecasesauth.GoogleSignInCallbackResult, err error) {
	const tag = tagPath + "/SignInCallback"

	if !u.config.Enabled {
		failure := domainmodelserror.NewError("Google authentication is disabled", domainmodelserror.Unimplemented, nil)
		u.logger.Error(ctx, tag, "Google callback is disabled", domainmodelslogger.Meta{"err": failure})
		return nil, failure
	}
	var result *domainusecasesauth.SessionResult

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

	if completed != nil && completed.Denied {
		if _, err := applicationshared.ReadAuthClient(ctx, u.client, u.space, completed.ClientId, completed.SpaceId, completed.RedirectUri); err != nil {
			u.logger.Error(ctx, tag, "failed to validate client after Google denial", domainmodelslogger.Meta{"err": err})
			return nil, err
		}
		u.logger.Warn(ctx, tag, "Google authorization was denied", nil)
		return &domainusecasesauth.GoogleSignInCallbackResult{RedirectUrl: redirectUrl(completed.RedirectUri, completed.ClientState, "error", "access_denied")}, nil
	}
	if completed == nil || completed.ClientId <= 0 || completed.SpaceId <= 0 || completed.RedirectUri == "" || completed.ClientState == "" || !completed.Identity.IsEmailVerified || completed.Identity.Subject == "" || completed.Identity.Email == "" {
		failure := domainmodelserror.NewError("invalid google identity", domainmodelserror.GoogleOAuthInvalid, nil)
		u.logger.Error(ctx, tag, "invalid google identity", domainmodelslogger.Meta{
			"err": failure,
		})

		return nil, failure
	}

	code, err := u.flow.Generate()
	if err != nil {
		u.logger.Error(ctx, tag, "failed to generate authentication handoff code", domainmodelslogger.Meta{"err": err})
		return nil, err
	}

	changed := false
	for attempt := 0; attempt < 2; attempt++ {
		result, changed = nil, false
		err = u.transactor.WithTx(ctx, func(txCtx context.Context) error {
			if _, err := applicationshared.ReadAuthClient(txCtx, u.client, u.space, completed.ClientId, completed.SpaceId, completed.RedirectUri); err != nil {
				u.logger.Error(txCtx, tag, "failed to validate Google callback client and redirect", domainmodelslogger.Meta{"err": err})
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
			if err := u.handoffs.Create(txCtx, domainmodelsentity.AuthHandoff{
				CodeHash:      u.flow.Hash(code),
				ClientId:      completed.ClientId,
				SpaceId:       completed.SpaceId,
				UserId:        result.User.Id,
				SpaceMemberId: result.SpaceMember.Id,
				RedirectUri:   completed.RedirectUri,
				IsNewUser:     newUser,
				ExpiresAt:     time.Now().Add(u.config.HandoffDuration),
			}); err != nil {
				u.logger.Error(txCtx, tag, "failed to persist authentication handoff", domainmodelslogger.Meta{"err": err})
				return err
			}

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

	return &domainusecasesauth.GoogleSignInCallbackResult{RedirectUrl: redirectUrl(completed.RedirectUri, completed.ClientState, "code", code)}, nil
}

func redirectUrl(uri, state, key, value string) string {
	query := url.Values{"state": {state}, key: {value}}
	return uri + "?" + query.Encode()
}
