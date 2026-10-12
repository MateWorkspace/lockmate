package presentationgrpcservermanagementpermission

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
	lockmatev1.UnimplementedPermissionServiceServer
	permission domainusecasesmanagement.Permission
}

func NewServer(permission domainusecasesmanagement.Permission) lockmatev1.PermissionServiceServer {
	return &server{permission: permission}
}

func (s *server) Create(ctx context.Context, req *lockmatev1.PermissionServiceCreateRequest) (*lockmatev1.PermissionServiceCreateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.permission.Create(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.CreatePermissionRequest{
		Slug:        req.GetSlug(),
		Name:        req.GetName(),
		Description: req.Description,
		Preferences: presentationgrpcmapper.OptionalPreferences(req.Preferences),
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.PermissionServiceCreateResponse{Id: id}, nil
}

func (s *server) ReadById(ctx context.Context, req *lockmatev1.PermissionServiceReadByIdRequest) (*lockmatev1.PermissionServiceReadByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.permission.ReadById(ctx, caller, req.GetSpaceId(), req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.PermissionServiceReadByIdResponse{Permission: presentationgrpcmapper.Permission(*result)}, nil
}

func (s *server) ReadBySlug(ctx context.Context, req *lockmatev1.PermissionServiceReadBySlugRequest) (*lockmatev1.PermissionServiceReadBySlugResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.permission.ReadBySlug(ctx, caller, req.GetSpaceId(), req.GetSlug())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.PermissionServiceReadBySlugResponse{Permission: presentationgrpcmapper.Permission(*result)}, nil
}

func (s *server) List(ctx context.Context, req *lockmatev1.PermissionServiceListRequest) (*lockmatev1.PermissionServiceListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.permission.ReadByFilter(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.FilterPermissionRequest{
		Page:   req.GetPage(),
		Limit:  req.GetLimit(),
		Search: req.Search,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.PermissionServiceListResponse{
		Data: presentationgrpcmapper.Permissions(page.Items),
		Page: &lockmatev1.PageResponse{Page: req.GetPage(), Limit: req.GetLimit(), TotalItems: page.Total},
	}, nil
}

func (s *server) Update(ctx context.Context, req *lockmatev1.PermissionServiceUpdateRequest) (*lockmatev1.PermissionServiceUpdateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.permission.UpdateById(ctx, caller, req.GetSpaceId(), req.GetId(), domainusecasesmanagement.UpdatePermissionRequest{
		Name:        req.Name,
		Description: req.Description,
		Preferences: presentationgrpcmapper.OptionalPreferences(req.Preferences),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.PermissionServiceUpdateResponse{}, nil
}

func (s *server) Delete(ctx context.Context, req *lockmatev1.PermissionServiceDeleteRequest) (*lockmatev1.PermissionServiceDeleteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.permission.DeleteById(ctx, caller, req.GetSpaceId(), req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.PermissionServiceDeleteResponse{}, nil
}
