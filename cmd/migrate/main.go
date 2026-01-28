package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/notification-system-moxicom/persistence-service/internal/config"
	"github.com/notification-system-moxicom/persistence-service/internal/repository"
)

func main() {
	var (
		command    = flag.String("command", "up", "Migration command: up, down, status, or version")
		dir        = flag.String("dir", "migrations", "Directory with migration files")
		dbURL      = flag.String("db-url", "", "Database connection URL (overrides config)")
		configPath = flag.String("config", "configs/config.yaml", "Path to config file")
		version    = flag.Int64("version", 0, "Version for up-to or down-to commands")
	)

	flag.Parse()

	dsn := *dbURL
	if dsn == "" {
		cfg, err := config.ReadConfig(*configPath)
		if err == nil {
			dsn = repository.BuildDSN(cfg.Connections.Postgres)
		}
	}

	if dsn == "" {
		log.Fatal("Database connection URL is required. Provide it via: -db-url flag, config file, or DATABASE_URL environment variable")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Warn("Failed to close database connection:", "error", err)
		}
	}()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	migrationsDir := *dir

	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for migrations directory: %v", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		log.Fatalf("Migrations directory does not exist: %s", absDir)
	}

	if !info.IsDir() {
		log.Fatalf("Path is not a directory: %s", absDir)
	}

	migrationsDir = absDir

	switch *command {
	case "up":
		if err := goose.Up(db, migrationsDir); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		log.Println("Migrations applied successfully")

	case "down":
		if err := goose.Down(db, migrationsDir); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}

		log.Println("Migrations rolled back successfully")

	case "down-to":
		if *version == 0 {
			log.Fatal("Version is required for down-to command")
		}

		if err := goose.DownTo(db, migrationsDir, *version); err != nil {
			log.Fatalf("Failed to rollback migrations to version %d: %v", *version, err)
		}

		log.Printf("Migrations rolled back to version %d successfully", *version)

	case "up-to":
		if *version == 0 {
			log.Fatal("Version is required for up-to command")
		}

		if err := goose.UpTo(db, migrationsDir, *version); err != nil {
			log.Fatalf("Failed to run migrations to version %d: %v", *version, err)
		}

		log.Printf("Migrations applied to version %d successfully", *version)

	case "status":
		if err := goose.Status(db, migrationsDir); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

	case "version":
		version, err := goose.GetDBVersion(db)
		if err != nil {
			log.Fatalf("Failed to get database version: %v", err)
		}

		log.Printf("Current database version: %d", version)

	case "create":
		if len(flag.Args()) == 0 {
			log.Fatal("Migration name is required for create command")
		}

		name := flag.Args()[0]
		if err := goose.Create(db, migrationsDir, name, "sql"); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}

		log.Printf("Migration created successfully: %s", name)

	default:
		log.Fatalf("Unknown command: %s. Available commands: up, down, down-to, up-to, status, version, create", *command)
	}
}
