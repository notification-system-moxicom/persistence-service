package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apperrors "github.com/notification-system-moxicom/persistence-service/internal/errors"
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
	system, err := s.repo.CreateSystem(ctx, request.GetName(), request.GetDescription())
	if err != nil {
		slog.Error("create system failed", "name", request.GetName(), "error", err)
		return nil, err
	}
	return system, nil
}

func (s *grpcService) GetSystems(ctx context.Context, request *rpcv1.GetSystemsRequest) (*rpcv1.Systems, error) {
	systems, err := s.repo.ListSystems(ctx)
	if err != nil {
		slog.Error("get systems failed", "error", err)
		return nil, err
	}

	return &rpcv1.Systems{Systems: systems}, nil
}

func (s *grpcService) AddUser(ctx context.Context, request *rpcv1.AddUserRequest) (*rpcv1.User, error) {
	user, err := s.repo.AddUser(ctx, request.GetSystemId(), request.GetIdAtSystem(), request.GetAdapters())
	if err != nil {
		slog.Error("add user failed", "system_id", request.GetSystemId(), "id_at_system", request.GetIdAtSystem(), "error", err)
		return nil, err
	}
	return user, nil
}

func (s *grpcService) GetUsers(ctx context.Context, request *rpcv1.GetUsersRequest) (*rpcv1.Users, error) {
	users, err := s.repo.ListUsers(ctx, request.GetSystemId())
	if err != nil {
		slog.Error("get users failed", "system_id", request.GetSystemId(), "error", err)
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

	system, err := s.repo.UpdateSystem(ctx, request.GetId(), name, description)
	if err != nil {
		slog.Error("update system failed", "id", request.GetId(), "error", err)
		return nil, err
	}
	return system, nil
}

func (s *grpcService) DeleteSystem(ctx context.Context, request *rpcv1.DeleteSystemRequest) (*rpcv1.InfoMessage, error) {
	if err := s.repo.DeleteSystem(ctx, request.GetId()); err != nil {
		slog.Error("delete system failed", "id", request.GetId(), "error", err)
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

	user, err := s.repo.UpdateUser(ctx, request.GetId(), idAtSystem, request.GetAdapters())
	if err != nil {
		slog.Error("update user failed", "id", request.GetId(), "error", err)
		return nil, err
	}
	return user, nil
}

func (s *grpcService) DeleteUser(ctx context.Context, request *rpcv1.DeleteUserRequest) (*rpcv1.InfoMessage, error) {
	if err := s.repo.DeleteUser(ctx, request.GetId()); err != nil {
		slog.Error("delete user failed", "id", request.GetId(), "error", err)
		return nil, err
	}

	return &rpcv1.InfoMessage{Message: "user deleted"}, nil
}

func (s *grpcService) Notify(ctx context.Context, request *rpcv1.NotifyRequest) (*rpcv1.NotifyResponse, error) {
	if err := validateNotifyRequest(request); err != nil {
		slog.Info("invalid notify request", "req", request)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	notificationID, err := s.repo.CreateNotification(
		ctx,
		request.GetSystemId(),
		request.GetUserIds(),
		request.GetContent(),
	)
	if err != nil {
		slog.Error("create notification failed", "system_id", request.GetSystemId(), "user_ids", request.GetUserIds(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create notification: %v", err)
	}

	return &rpcv1.NotifyResponse{NotificationId: notificationID}, nil
}

func validateNotifyRequest(req *rpcv1.NotifyRequest) error {
	var details []string

	if req.GetSystemId() == "" {
		details = append(details, "system_id is required")
	} else if _, err := uuid.Parse(req.GetSystemId()); err != nil {
		details = append(details, "system_id must be a valid UUID")
	}

	if len(req.GetUserIds()) == 0 {
		details = append(details, "user_ids must not be empty")
	}

	if req.GetContent() == "" {
		details = append(details, "content is required")
	}

	if len(details) > 0 {
		return apperrors.NewValidationError("invalid notify request", details...)
	}

	return nil
}

func (s *grpcService) Ping(ctx context.Context) error {
	return nil
}

func NewService(repo repository.Repository) Service {
	return &grpcService{
		repo: repo,
	}
}
