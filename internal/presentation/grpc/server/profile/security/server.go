package presentationgrpcserverprofilesecurity

import (
	"context"

	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
	presentationgrpcmapper "github.com/MateWorkspace/lockmate/internal/presentation/grpc/mapper"
	presentationgrpcservershared "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	lockmatev1.UnimplementedSecurityServiceServer
	security domainusecasesprofile.Security
}

func NewServer(security domainusecasesprofile.Security) lockmatev1.SecurityServiceServer {
	return &server{security: security}
}

func (s *server) GetAuthMethods(ctx context.Context, req *lockmatev1.SecurityServiceGetAuthMethodsRequest) (*lockmatev1.SecurityServiceGetAuthMethodsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.security.GetAuthMethods(ctx, caller)
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.SecurityServiceGetAuthMethodsResponse{Methods: presentationgrpcmapper.AuthMethods(*result)}, nil
}

func (s *server) UpdateClassicAuth(ctx context.Context, req *lockmatev1.SecurityServiceUpdateClassicAuthRequest) (*lockmatev1.SecurityServiceUpdateClassicAuthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.security.UpdateClassicAuth(ctx, caller, domainusecasesprofile.UpdateClassicAuthRequest{
		Username:     req.Username,
		Email:        req.Email,
		EmailSetNull: req.GetEmailSetNull(),
		Phone:        req.Phone,
		PhoneSetNull: req.GetPhoneSetNull(),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SecurityServiceUpdateClassicAuthResponse{}, nil
}

func (s *server) ChangePassword(ctx context.Context, req *lockmatev1.SecurityServiceChangePasswordRequest) (*lockmatev1.SecurityServiceChangePasswordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.security.ChangePassword(ctx, caller, domainusecasesprofile.ChangePasswordRequest{
		CurrentPassword: req.GetCurrentPassword(),
		NewPassword:     req.GetNewPassword(),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SecurityServiceChangePasswordResponse{}, nil
}
