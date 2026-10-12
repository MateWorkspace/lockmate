package presentationgrpcserverauthhandoff

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
	lockmatev1.UnimplementedHandoffServiceServer
	handoff domainusecasesauth.Handoff
}

func NewServer(handoff domainusecasesauth.Handoff) lockmatev1.HandoffServiceServer {
	return &server{handoff: handoff}
}

func (s *server) Exchange(ctx context.Context, req *lockmatev1.HandoffServiceExchangeRequest) (*lockmatev1.HandoffServiceExchangeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	client, err := presentationgrpcservershared.Client(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.handoff.Exchange(ctx, client, domainusecasesauth.ExchangeAuthHandoffRequest{Code: req.GetCode()})
	if err != nil {
		return nil, presentationgrpcmapper.Error(err)
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &lockmatev1.HandoffServiceExchangeResponse{Token: presentationgrpcmapper.Token(*result)}, nil
}
