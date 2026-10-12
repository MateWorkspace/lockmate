package presentationgrpcserverauthsession

import (
	"context"

	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	presentationgrpcmapper "github.com/MateWorkspace/lockmate/internal/presentation/grpc/mapper"
	presentationgrpcservershared "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	lockmatev1.UnimplementedSessionServiceServer
	session domainusecasesauth.Session
}

func NewServer(session domainusecasesauth.Session) lockmatev1.SessionServiceServer {
	return &server{session: session}
}

func (s *server) SignIn(ctx context.Context, req *lockmatev1.SessionServiceSignInRequest) (*lockmatev1.SessionServiceSignInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	client, err := presentationgrpcservershared.Client(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.session.SignIn(ctx, client, domainusecasesauth.SignInRequest{Username: req.GetUsername(), Password: req.GetPassword()})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.SessionServiceSignInResponse{Token: presentationgrpcmapper.Token(*result)}, nil
}

func (s *server) Register(ctx context.Context, req *lockmatev1.SessionServiceRegisterRequest) (*lockmatev1.SessionServiceRegisterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	client, err := presentationgrpcservershared.Client(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.session.Register(ctx, client, domainusecasesauth.RegisterRequest{
		Name:     req.GetName(),
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Bio:      req.Bio,
		Email:    req.Email,
		Phone:    req.Phone,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.SessionServiceRegisterResponse{Token: presentationgrpcmapper.Token(*result)}, nil
}

func (s *server) Refresh(ctx context.Context, req *lockmatev1.SessionServiceRefreshRequest) (*lockmatev1.SessionServiceRefreshResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	client, err := presentationgrpcservershared.Client(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.session.Refresh(ctx, client, domainusecasesauth.RefreshRequest{RefreshToken: req.GetRefreshToken()})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.SessionServiceRefreshResponse{Token: presentationgrpcmapper.Token(*result)}, nil
}
