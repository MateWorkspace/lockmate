package presentationgrpcservermanagementrole

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
	lockmatev1.UnimplementedRoleServiceServer
	role domainusecasesmanagement.Role
}

func NewServer(role domainusecasesmanagement.Role) lockmatev1.RoleServiceServer {
	return &server{role: role}
}

func (s *server) Create(ctx context.Context, req *lockmatev1.RoleServiceCreateRequest) (*lockmatev1.RoleServiceCreateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.role.Create(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.CreateRoleRequest{
		Slug:        req.GetSlug(),
		Name:        req.GetName(),
		Description: req.Description,
		IsDefault:   req.IsDefault,
		Preferences: presentationgrpcmapper.OptionalPreferences(req.Preferences),
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.RoleServiceCreateResponse{Id: id}, nil
}

func (s *server) ReadById(ctx context.Context, req *lockmatev1.RoleServiceReadByIdRequest) (*lockmatev1.RoleServiceReadByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.role.ReadById(ctx, caller, req.GetSpaceId(), req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.RoleServiceReadByIdResponse{Role: presentationgrpcmapper.Role(*result)}, nil
}

func (s *server) ReadBySlug(ctx context.Context, req *lockmatev1.RoleServiceReadBySlugRequest) (*lockmatev1.RoleServiceReadBySlugResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.role.ReadBySlug(ctx, caller, req.GetSpaceId(), req.GetSlug())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.RoleServiceReadBySlugResponse{Role: presentationgrpcmapper.Role(*result)}, nil
}

func (s *server) List(ctx context.Context, req *lockmatev1.RoleServiceListRequest) (*lockmatev1.RoleServiceListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.role.ReadByFilter(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.FilterRoleRequest{
		Page:      req.GetPage(),
		Limit:     req.GetLimit(),
		Search:    req.Search,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.RoleServiceListResponse{
		Data: presentationgrpcmapper.Roles(page.Items),
		Page: &lockmatev1.PageResponse{Page: req.GetPage(), Limit: req.GetLimit(), TotalItems: page.Total},
	}, nil
}

func (s *server) ReadDefault(ctx context.Context, req *lockmatev1.RoleServiceReadDefaultRequest) (*lockmatev1.RoleServiceReadDefaultResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.role.ReadDefault(ctx, caller, req.GetSpaceId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.RoleServiceReadDefaultResponse{Role: presentationgrpcmapper.Role(*result)}, nil
}

func (s *server) Update(ctx context.Context, req *lockmatev1.RoleServiceUpdateRequest) (*lockmatev1.RoleServiceUpdateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.role.UpdateById(ctx, caller, req.GetSpaceId(), req.GetId(), domainusecasesmanagement.UpdateRoleRequest{
		Name:        req.Name,
		Description: req.Description,
		IsDefault:   req.IsDefault,
		Preferences: presentationgrpcmapper.OptionalPreferences(req.Preferences),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.RoleServiceUpdateResponse{}, nil
}

func (s *server) SetDefault(ctx context.Context, req *lockmatev1.RoleServiceSetDefaultRequest) (*lockmatev1.RoleServiceSetDefaultResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.role.SetDefault(ctx, caller, req.GetSpaceId(), req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.RoleServiceSetDefaultResponse{}, nil
}

func (s *server) Delete(ctx context.Context, req *lockmatev1.RoleServiceDeleteRequest) (*lockmatev1.RoleServiceDeleteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.role.DeleteById(ctx, caller, req.GetSpaceId(), req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.RoleServiceDeleteResponse{}, nil
}
