package presentationgrpcserverauthgoogle

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
	lockmatev1.UnimplementedGoogleServiceServer
	google domainusecasesauth.Google
}

func NewServer(google domainusecasesauth.Google) lockmatev1.GoogleServiceServer {
	return &server{google: google}
}

func (s *server) Start(ctx context.Context, req *lockmatev1.GoogleServiceStartRequest) (*lockmatev1.GoogleServiceStartResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	client, err := presentationgrpcservershared.Client(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.google.Start(ctx, client, domainusecasesauth.StartGoogleSignInRequest{State: req.GetState()})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.GoogleServiceStartResponse{LaunchUrl: result.LaunchUrl}, nil
}
