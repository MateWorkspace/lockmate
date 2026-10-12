package presentationgrpcservermanagementuser

import (
	"context"

	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	presentationgrpcmapper "github.com/MateWorkspace/lockmate/internal/presentation/grpc/mapper"
	presentationgrpcservershared "github.com/MateWorkspace/lockmate/internal/presentation/grpc/server/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	lockmatev1.UnimplementedUserServiceServer
	user domainusecasesmanagement.User
}

func NewServer(user domainusecasesmanagement.User) lockmatev1.UserServiceServer {
	return &server{user: user}
}

func (s *server) Create(ctx context.Context, req *lockmatev1.UserServiceCreateRequest) (*lockmatev1.UserServiceCreateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.user.Create(ctx, caller, domainusecasesmanagement.CreateUserRequest{
		Name:            req.GetName(),
		Bio:             req.Bio,
		AvatarPath:      req.AvatarPath,
		Preferences:     presentationgrpcmapper.OptionalPreferences(req.Preferences),
		Username:        req.GetUsername(),
		Password:        req.GetPassword(),
		Email:           req.Email,
		Phone:           req.Phone,
		IsEmailVerified: req.IsEmailVerified,
		IsPhoneVerified: req.IsPhoneVerified,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.UserServiceCreateResponse{Id: id}, nil
}

func (s *server) ReadById(ctx context.Context, req *lockmatev1.UserServiceReadByIdRequest) (*lockmatev1.UserServiceReadByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.user.ReadById(ctx, caller, req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.UserServiceReadByIdResponse{User: presentationgrpcmapper.User(*result)}, nil
}

func (s *server) List(ctx context.Context, req *lockmatev1.UserServiceListRequest) (*lockmatev1.UserServiceListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.user.ReadByFilter(ctx, caller, domainusecasesmanagement.FilterUserRequest{
		Page:   req.GetPage(),
		Limit:  req.GetLimit(),
		Search: req.Search,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.UserServiceListResponse{
		Data: presentationgrpcmapper.Users(page.Items),
		Page: &lockmatev1.PageResponse{Page: req.GetPage(), Limit: req.GetLimit(), TotalItems: page.Total},
	}, nil
}

func (s *server) Update(ctx context.Context, req *lockmatev1.UserServiceUpdateRequest) (*lockmatev1.UserServiceUpdateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.user.UpdateById(ctx, caller, req.GetId(), domainusecasesmanagement.UpdateUserRequest{
		Name:              req.Name,
		Bio:               req.Bio,
		AvatarPath:        req.AvatarPath,
		AvatarPathSetNull: req.GetAvatarPathSetNull(),
		Preferences:       presentationgrpcmapper.OptionalPreferences(req.Preferences),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.UserServiceUpdateResponse{}, nil
}

func (s *server) Delete(ctx context.Context, req *lockmatev1.UserServiceDeleteRequest) (*lockmatev1.UserServiceDeleteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.user.DeleteById(ctx, caller, req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.UserServiceDeleteResponse{}, nil
}
