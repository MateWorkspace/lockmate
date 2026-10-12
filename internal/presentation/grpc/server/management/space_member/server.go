package presentationgrpcservermanagementspacemember

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
	lockmatev1.UnimplementedSpaceMemberServiceServer
	spaceMember domainusecasesmanagement.SpaceMember
}

func NewServer(spaceMember domainusecasesmanagement.SpaceMember) lockmatev1.SpaceMemberServiceServer {
	return &server{spaceMember: spaceMember}
}

func (s *server) Create(ctx context.Context, req *lockmatev1.SpaceMemberServiceCreateRequest) (*lockmatev1.SpaceMemberServiceCreateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.spaceMember.Create(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.CreateSpaceMemberRequest{
		UserId:      req.GetUserId(),
		IsActive:    req.IsActive,
		Preferences: presentationgrpcmapper.OptionalPreferences(req.Preferences),
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberServiceCreateResponse{Id: id}, nil
}

func (s *server) ReadById(ctx context.Context, req *lockmatev1.SpaceMemberServiceReadByIdRequest) (*lockmatev1.SpaceMemberServiceReadByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.spaceMember.ReadById(ctx, caller, req.GetSpaceId(), req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.SpaceMemberServiceReadByIdResponse{SpaceMember: presentationgrpcmapper.SpaceMember(*result)}, nil
}

func (s *server) ReadByUserId(ctx context.Context, req *lockmatev1.SpaceMemberServiceReadByUserIdRequest) (*lockmatev1.SpaceMemberServiceReadByUserIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.spaceMember.ReadByUserId(ctx, caller, req.GetSpaceId(), req.GetUserId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.SpaceMemberServiceReadByUserIdResponse{SpaceMember: presentationgrpcmapper.SpaceMember(*result)}, nil
}

func (s *server) List(ctx context.Context, req *lockmatev1.SpaceMemberServiceListRequest) (*lockmatev1.SpaceMemberServiceListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.spaceMember.ReadByFilter(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.FilterSpaceMemberRequest{
		Page:     req.GetPage(),
		Limit:    req.GetLimit(),
		Search:   req.Search,
		UserId:   req.UserId,
		RoleId:   req.RoleId,
		IsActive: req.IsActive,
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberServiceListResponse{
		Data: presentationgrpcmapper.SpaceMemberWithUsers(page.Items),
		Page: &lockmatev1.PageResponse{Page: req.GetPage(), Limit: req.GetLimit(), TotalItems: page.Total},
	}, nil
}

func (s *server) Update(ctx context.Context, req *lockmatev1.SpaceMemberServiceUpdateRequest) (*lockmatev1.SpaceMemberServiceUpdateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.spaceMember.UpdateById(ctx, caller, req.GetSpaceId(), req.GetId(), domainusecasesmanagement.UpdateSpaceMemberRequest{
		IsActive:    req.IsActive,
		Preferences: presentationgrpcmapper.OptionalPreferences(req.Preferences),
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberServiceUpdateResponse{}, nil
}

func (s *server) ReadEffectiveRoles(ctx context.Context, req *lockmatev1.SpaceMemberServiceReadEffectiveRolesRequest) (*lockmatev1.SpaceMemberServiceReadEffectiveRolesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	roles, err := s.spaceMember.ReadEffectiveRolesBySpaceMemberId(ctx, caller, req.GetSpaceId(), req.GetSpaceMemberId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberServiceReadEffectiveRolesResponse{Roles: presentationgrpcmapper.Roles(roles)}, nil
}

func (s *server) ReadEffectivePermissions(ctx context.Context, req *lockmatev1.SpaceMemberServiceReadEffectivePermissionsRequest) (*lockmatev1.SpaceMemberServiceReadEffectivePermissionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	permissions, err := s.spaceMember.ReadEffectivePermissionsBySpaceMemberId(ctx, caller, req.GetSpaceId(), req.GetSpaceMemberId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberServiceReadEffectivePermissionsResponse{Permissions: presentationgrpcmapper.Permissions(permissions)}, nil
}

func (s *server) Delete(ctx context.Context, req *lockmatev1.SpaceMemberServiceDeleteRequest) (*lockmatev1.SpaceMemberServiceDeleteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.spaceMember.DeleteById(ctx, caller, req.GetSpaceId(), req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.SpaceMemberServiceDeleteResponse{}, nil
}
