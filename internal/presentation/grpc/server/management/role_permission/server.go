package presentationgrpcservermanagementrolepermission

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
	lockmatev1.UnimplementedRolePermissionServiceServer
	rolePermission domainusecasesmanagement.RolePermission
}

func NewServer(rolePermission domainusecasesmanagement.RolePermission) lockmatev1.RolePermissionServiceServer {
	return &server{rolePermission: rolePermission}
}

func (s *server) Create(ctx context.Context, req *lockmatev1.RolePermissionServiceCreateRequest) (*lockmatev1.RolePermissionServiceCreateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.rolePermission.Create(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.CreateRolePermissionRequest{
		RoleId:       req.GetRoleId(),
		PermissionId: req.GetPermissionId(),
	})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.RolePermissionServiceCreateResponse{Id: id}, nil
}

func (s *server) ReadById(ctx context.Context, req *lockmatev1.RolePermissionServiceReadByIdRequest) (*lockmatev1.RolePermissionServiceReadByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.rolePermission.ReadById(ctx, caller, req.GetSpaceId(), req.GetId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.RolePermissionServiceReadByIdResponse{Details: presentationgrpcmapper.RolePermissionDetails(*result)}, nil
}

func (s *server) ReadByPair(ctx context.Context, req *lockmatev1.RolePermissionServiceReadByPairRequest) (*lockmatev1.RolePermissionServiceReadByPairResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.rolePermission.ReadByRoleIdAndPermissionId(ctx, caller, req.GetSpaceId(), req.GetRoleId(), req.GetPermissionId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.RolePermissionServiceReadByPairResponse{Details: presentationgrpcmapper.RolePermissionDetails(*result)}, nil
}

func (s *server) ReadByRoleId(ctx context.Context, req *lockmatev1.RolePermissionServiceReadByRoleIdRequest) (*lockmatev1.RolePermissionServiceReadByRoleIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.rolePermission.ReadByRoleId(ctx, caller, req.GetSpaceId(), req.GetRoleId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.RolePermissionServiceReadByRoleIdResponse{Data: presentationgrpcmapper.RolePermissionWithPermissions(result)}, nil
}

func (s *server) ReadByPermissionId(ctx context.Context, req *lockmatev1.RolePermissionServiceReadByPermissionIdRequest) (*lockmatev1.RolePermissionServiceReadByPermissionIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.rolePermission.ReadByPermissionId(ctx, caller, req.GetSpaceId(), req.GetPermissionId())
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.RolePermissionServiceReadByPermissionIdResponse{Data: presentationgrpcmapper.RolePermissionWithRoles(result)}, nil
}

func (s *server) Delete(ctx context.Context, req *lockmatev1.RolePermissionServiceDeleteRequest) (*lockmatev1.RolePermissionServiceDeleteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.rolePermission.DeleteById(ctx, caller, req.GetSpaceId(), req.GetId()); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.RolePermissionServiceDeleteResponse{}, nil
}

func (s *server) DeleteBySelectors(ctx context.Context, req *lockmatev1.RolePermissionServiceDeleteBySelectorsRequest) (*lockmatev1.RolePermissionServiceDeleteBySelectorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	caller, err := presentationgrpcservershared.Claims(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.rolePermission.DeleteByRoleIdOrPermissionId(ctx, caller, req.GetSpaceId(), domainusecasesmanagement.DeleteRolePermissionRequest{
		RoleId:       req.RoleId,
		PermissionId: req.PermissionId,
	}); err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}

	return &lockmatev1.RolePermissionServiceDeleteBySelectorsResponse{}, nil
}
