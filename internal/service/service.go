package service

import (
	"context"

	"github.com/notification-system-moxicom/persistence-service/internal/repository"
	rpcv1 "github.com/notification-system-moxicom/persistence-service/pkg/proto/gen/persistence/v1"
)

type Service interface {
	Ping(ctx context.Context) error
}

type grpcService struct {
	rpcv1.UnimplementedPersistenceServiceServer

	repo repository.Repository
}

func (s *grpcService) CreateSystem(ctx context.Context, request *rpcv1.CreateSystemRequest) (*rpcv1.System, error) {
	return s.repo.CreateSystem(ctx, request.GetName(), request.GetDescription())
}

func (s *grpcService) GetSystems(ctx context.Context, request *rpcv1.GetSystemsRequest) (*rpcv1.Systems, error) {
	systems, err := s.repo.ListSystems(ctx)
	if err != nil {
		return nil, err
	}

	return &rpcv1.Systems{Systems: systems}, nil
}

func (s *grpcService) AddUser(ctx context.Context, request *rpcv1.AddUserRequest) (*rpcv1.User, error) {
	return s.repo.AddUser(ctx, request.GetSystemId(), request.GetIdAtSystem(), request.GetAdapters())
}

func (s *grpcService) GetUsers(ctx context.Context, request *rpcv1.GetUsersRequest) (*rpcv1.Users, error) {
	users, err := s.repo.ListUsers(ctx, request.GetSystemId())
	if err != nil {
		return nil, err
	}

	return &rpcv1.Users{Users: users}, nil
}

func (s *grpcService) Notify(ctx context.Context, request *rpcv1.NotifyRequest) (*rpcv1.InfoMessage, error) {
	return &rpcv1.InfoMessage{Message: "notification accepted"}, nil
}

func (s *grpcService) Ping(ctx context.Context) error {
	return nil
}

func NewService(repo repository.Repository) *grpcService {
	return &grpcService{
		repo: repo,
	}
}
