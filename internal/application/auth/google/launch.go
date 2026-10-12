package applicationauthgoogle

import (
	"context"
	"net/url"
	"regexp"
	"time"

	applicationshared "github.com/MateWorkspace/lockmate/internal/application/shared"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
)

var projectStatePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{32,512}$`)

func (u *usecase) Start(ctx context.Context, client *domainmodelsentity.ClientIdentity, request domainusecasesauth.StartGoogleSignInRequest) (*domainusecasesauth.StartGoogleSignInResult, error) {
	const tag = tagPath + "/Start"

	if !u.config.Enabled {
		err := domainmodelserror.NewError("Google authentication is disabled", domainmodelserror.Unimplemented, nil)
		u.logger.Error(ctx, tag, "Google sign-in initiation is disabled", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	if !projectStatePattern.MatchString(request.State) {
		err := domainmodelserror.NewError("state must contain 32 to 512 URL-safe characters", domainmodelserror.BadArgs, nil)
		u.logger.Error(ctx, tag, "invalid project authentication correlation state", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	live, err := applicationshared.ReadClient(ctx, u.client, u.space, client)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to validate Google initiation client", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	if live.AuthRedirectUri == nil {
		err := domainmodelserror.NewError("client authentication redirect is not configured", domainmodelserror.BadState, nil)
		u.logger.Error(ctx, tag, "client is missing its authentication redirect", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	if err := u.validator.ClientAuthRedirectUri(*live.AuthRedirectUri); err != nil {
		u.logger.Error(ctx, tag, "client authentication redirect is invalid", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	ticket, err := u.flow.Generate()
	if err != nil {
		u.logger.Error(ctx, tag, "failed to generate Google browser launch ticket", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	if err := u.launches.Create(ctx, domainmodelsentity.GoogleOAuthLaunch{
		TicketHash:  u.flow.Hash(ticket),
		ClientId:    live.Id,
		SpaceId:     live.SpaceId,
		RedirectUri: *live.AuthRedirectUri,
		ClientState: request.State,
		ExpiresAt:   time.Now().Add(u.config.LaunchDuration),
	}); err != nil {
		u.logger.Error(ctx, tag, "failed to persist Google browser launch ticket", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	return &domainusecasesauth.StartGoogleSignInResult{LaunchUrl: u.config.PublicUrl + "/api/v1/auth/google/launch?ticket=" + url.QueryEscape(ticket)}, nil
}

func (u *usecase) Launch(ctx context.Context, ticket string) (*domainusecasesauth.LaunchGoogleSignInResult, error) {
	const tag = tagPath + "/Launch"

	if !u.config.Enabled {
		err := domainmodelserror.NewError("Google authentication is disabled", domainmodelserror.Unimplemented, nil)
		u.logger.Error(ctx, tag, "Google browser launch is disabled", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	if len(ticket) != 43 {
		err := domainmodelserror.NewError("invalid browser launch ticket", domainmodelserror.GoogleOAuthStateInvalid, nil)
		u.logger.Error(ctx, tag, "invalid Google browser launch ticket", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	launch, err := u.launches.Consume(ctx, u.flow.Hash(ticket))
	if err != nil {
		u.logger.Error(ctx, tag, "failed to consume Google browser launch ticket", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	if launch == nil {
		err := domainmodelserror.NewError("browser launch record is missing", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "Google browser launch record is missing", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	if _, err := applicationshared.ReadAuthClient(ctx, u.client, u.space, launch.ClientId, launch.SpaceId, launch.RedirectUri); err != nil {
		u.logger.Error(ctx, tag, "failed to validate Google browser launch scope", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	binding, err := u.flow.Generate()
	if err != nil {
		u.logger.Error(ctx, tag, "failed to generate Google browser binding", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	started, err := u.oauth.Begin(ctx, domaincontractsutility.BeginGoogleOAuthArgs{
		ClientId:    launch.ClientId,
		SpaceId:     launch.SpaceId,
		RedirectUri: launch.RedirectUri,
		ClientState: launch.ClientState,
		Binding:     binding,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to begin Google provider authorization", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	if started == nil || started.State == "" || started.AuthorizationUrl == "" {
		err := domainmodelserror.NewError("Google authorization result is missing", domainmodelserror.Failure, nil)
		u.logger.Error(ctx, tag, "Google provider authorization result is missing", domainmodelslogger.Meta{"err": err})

		return nil, err
	}

	return &domainusecasesauth.LaunchGoogleSignInResult{
		AuthorizationUrl: started.AuthorizationUrl,
		State:            started.State,
		Binding:          binding,
		CookieDuration:   int(u.config.AttemptDuration / time.Second),
	}, nil
}
