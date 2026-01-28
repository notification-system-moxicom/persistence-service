package service

import (
	"context"

	"github.com/notification-system-moxicom/persistence-service/internal/repository"
	rpcv1 "github.com/notification-system-moxicom/persistence-service/pkg/proto/gen/persistence/v1"
)

type Service interface {
	Ping(ctx context.Context) error
	rpcv1.PersistenceServiceServer
}

type grpcService struct {
	repo repository.Repository
	rpcv1.UnimplementedPersistenceServiceServer
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

func (s *grpcService) UpdateSystem(ctx context.Context, request *rpcv1.UpdateSystemRequest) (*rpcv1.System, error) {
	var name, description *string
	if request.GetName() != "" {
		n := request.GetName()
		name = &n
	}

	if request.GetDescription() != "" {
		d := request.GetDescription()
		description = &d
	}

	return s.repo.UpdateSystem(ctx, request.GetId(), name, description)
}

func (s *grpcService) DeleteSystem(ctx context.Context, request *rpcv1.DeleteSystemRequest) (*rpcv1.InfoMessage, error) {
	if err := s.repo.DeleteSystem(ctx, request.GetId()); err != nil {
		return nil, err
	}

	return &rpcv1.InfoMessage{Message: "system deleted"}, nil
}

func (s *grpcService) UpdateUser(ctx context.Context, request *rpcv1.UpdateUserRequest) (*rpcv1.User, error) {
	var idAtSystem *string
	if request.GetIdAtSystem() != "" {
		id := request.GetIdAtSystem()
		idAtSystem = &id
	}

	return s.repo.UpdateUser(ctx, request.GetId(), idAtSystem, request.GetAdapters())
}

func (s *grpcService) DeleteUser(ctx context.Context, request *rpcv1.DeleteUserRequest) (*rpcv1.InfoMessage, error) {
	if err := s.repo.DeleteUser(ctx, request.GetId()); err != nil {
		return nil, err
	}

	return &rpcv1.InfoMessage{Message: "user deleted"}, nil
}

func (s *grpcService) Notify(ctx context.Context, request *rpcv1.NotifyRequest) (*rpcv1.InfoMessage, error) {
	return &rpcv1.InfoMessage{Message: "notification accepted"}, nil
}

func (s *grpcService) Ping(ctx context.Context) error {
	return nil
}

func NewService(repo repository.Repository) Service {
	return &grpcService{
		repo: repo,
	}
}
