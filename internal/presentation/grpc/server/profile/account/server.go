package presentationgrpcserverprofileaccount

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
	lockmatev1.UnimplementedAccountServiceServer
	account domainusecasesprofile.Account
}

func NewServer(account domainusecasesprofile.Account) lockmatev1.AccountServiceServer {
	return &server{account: account}
}

func (s *server) GetProfile(ctx context.Context, req *lockmatev1.AccountServiceGetProfileRequest) (*lockmatev1.AccountServiceGetProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.account.GetProfile(ctx, caller)
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.AccountServiceGetProfileResponse{Profile: presentationgrpcmapper.Profile(*result)}, nil
}

func (s *server) GetRoles(ctx context.Context, req *lockmatev1.AccountServiceGetRolesRequest) (*lockmatev1.AccountServiceGetRolesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	roles, err := s.account.GetRoles(ctx, caller)
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.AccountServiceGetRolesResponse{Roles: presentationgrpcmapper.Roles(roles)}, nil
}

func (s *server) GetPermissions(ctx context.Context, req *lockmatev1.AccountServiceGetPermissionsRequest) (*lockmatev1.AccountServiceGetPermissionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	permissions, err := s.account.GetPermissions(ctx, caller)
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.AccountServiceGetPermissionsResponse{Permissions: presentationgrpcmapper.Permissions(permissions)}, nil
}

func (s *server) UpdateProfile(ctx context.Context, req *lockmatev1.AccountServiceUpdateProfileRequest) (*lockmatev1.AccountServiceUpdateProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.account.UpdateProfile(ctx, caller, domainusecasesprofile.UpdateProfileRequest{
		Name:              req.Name,
		Bio:               req.Bio,
		AvatarPath:        req.AvatarPath,
		AvatarPathSetNull: req.GetAvatarPathSetNull(),
		Preferences:       presentationgrpcmapper.OptionalPreferences(req.Preferences),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.AccountServiceUpdateProfileResponse{}, nil
}

func (s *server) UpdateMembershipPreferences(ctx context.Context, req *lockmatev1.AccountServiceUpdateMembershipPreferencesRequest) (*lockmatev1.AccountServiceUpdateMembershipPreferencesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.account.UpdateMembershipPreferences(ctx, caller, domainusecasesprofile.UpdateMembershipPreferencesRequest{
		Preferences: presentationgrpcmapper.RequiredPreferences(req.GetPreferences()),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.AccountServiceUpdateMembershipPreferencesResponse{}, nil
}
