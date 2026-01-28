package service

import (
	"context"

	rpcv1 "github.com/notification-system-moxicom/persistence-service/pkg/proto/gen/persistence/v1"
)

// Service is the internal abstraction used by other components (like readiness checks).
// It only exposes the methods they actually need.
type Service interface {
	Ping(ctx context.Context) error
}

// grpcService is the concrete gRPC implementation that also satisfies the generated
// rpcv1.PersistenceServiceServer interface.
type grpcService struct {
	rpcv1.UnimplementedPersistenceServiceServer
}

func (s *grpcService) CreateSystem(ctx context.Context, request *rpcv1.CreateSystemRequest) (*rpcv1.System, error) {
	//TODO implement me
	panic("implement me")
}

func (s *grpcService) GetSystems(ctx context.Context, request *rpcv1.GetSystemsRequest) (*rpcv1.Systems, error) {
	//TODO implement me
	panic("implement me")
}

func (s *grpcService) AddUser(ctx context.Context, request *rpcv1.AddUserRequest) (*rpcv1.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s *grpcService) GetUsers(ctx context.Context, request *rpcv1.GetUsersRequest) (*rpcv1.Users, error) {
	//TODO implement me
	panic("implement me")
}

func (s *grpcService) Notify(ctx context.Context, request *rpcv1.NotifyRequest) (*rpcv1.InfoMessage, error) {
	//TODO implement me
	panic("implement me")
}

func (s *grpcService) Ping(ctx context.Context) error {
	return nil
}

// NewService constructs the concrete gRPC service implementation.
// It returns the concrete type so callers can use it both as:
//   - rpcv1.PersistenceServiceServer (for gRPC registration)
//   - service.Service (for internal use, e.g. readiness checks)
func NewService(_ any) *grpcService {
	return &grpcService{}
}
