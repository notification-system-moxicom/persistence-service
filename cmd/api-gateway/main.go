package main

import (
	"context"
	"flag"
	"log/slog"

	"github.com/notification-system-moxicom/persistence-service/internal/config"
	"github.com/notification-system-moxicom/persistence-service/internal/kafka"
	"github.com/notification-system-moxicom/persistence-service/internal/repository"
	"github.com/notification-system-moxicom/persistence-service/internal/rpc"
	"github.com/notification-system-moxicom/persistence-service/internal/validation"
	"github.com/notification-system-moxicom/persistence-service/pkg/logger"
)

func main() {
	var configPath string

	logger.SetDefaults(nil)
	flag.StringVar(&configPath, "c", "config.yaml", "Set path to config file.")
	flag.Parse()

	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		slog.Error("can't configure from config file:", slog.String("error", err.Error()))
		return
	}

	schemaFiles := map[string]string{
		"notification_message": "schemas/send_notification.json",
	}

	validator, err := validation.NewJSONSchemaMessageValidator(schemaFiles)
	if err != nil {
		slog.Error("failed to create JSON schema validator:", slog.String("error", err.Error()))
		return
	}

	// Initialize PostgreSQL connection pool.
	ctx := context.Background()

	repo, err := repository.New(ctx, cfg.Connections.Postgres)
	if err != nil {
		slog.Error("failed to create repository:", slog.String("error", err.Error()))
		return
	}

	_, err = kafka.NewService(&cfg.Connections.Kafka.CamundaCore, validator)
	if err != nil {
		slog.Error("failed to create Kafka service:", slog.String("error", err.Error()))

		return
	}

	rpcServ := rpc.NewGRPC(&cfg.Server.GRPC, repo)
	if err = rpcServ.Listen(); err != nil {
		slog.Error("failed to listen RPC server", slog.String("error", err.Error()))
		return
	}

	slog.Info("finishing")
}
