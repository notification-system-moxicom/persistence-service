package rpc

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/notification-system-moxicom/persistence-service/internal/service"
	rpcv1 "github.com/notification-system-moxicom/persistence-service/pkg/proto/gen/persistence/v1"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
)

type GRPCConfig struct {
	Port string `yaml:"port"`
}

type GRPC struct {
	config *GRPCConfig
	s      *grpc.Server
	svc    service.Service
}

func (s *GRPC) Alive(_ http.ResponseWriter, _ *http.Request, _ map[string]string) {}

func (s *GRPC) Ready(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	ctx, cancel := context.WithDeadline(r.Context(), time.Now().Add(2*time.Second))
	defer cancel()

	if err := s.svc.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func NewGRPC(
	config *GRPCConfig,
	rep any, // TODO fixme
) *GRPC {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcValidator.UnaryServerInterceptor(),
		),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)

	functionSrv := service.NewService(rep)

	rpcv1.RegisterPersistenceServiceServer(grpcServer, functionSrv)

	return &GRPC{
		config: config,
		s:      grpcServer,
		svc:    functionSrv,
	}
}

func (s *GRPC) Listen() error {
	lis, err := net.Listen("tcp", s.config.Port)
	if err != nil {
		return err
	}

	slog.Info("grpc server started", "port", s.config.Port)

	err = view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return err
	}

	err = s.s.Serve(lis)
	if err != nil {
		return err
	}

	return nil
}
