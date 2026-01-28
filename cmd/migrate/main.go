package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
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
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		command    = flag.String("command", "up", "Migration command: up, down, status, or version")
		dir        = flag.String("dir", "migrations", "Directory with migration files")
		configPath = flag.String("config", "configs/config.yaml", "Path to config file")
		version    = flag.Int64("version", 0, "Version for up-to or down-to commands")
	)

	flag.Parse()

	cfg, err := config.ReadConfig(*configPath)
	if err != nil {
		return err
	}

	dsn := repository.BuildDSN(cfg.Connections.Postgres)
	if dsn == "" {
		return errors.New("dsn is nil")
	}

	db, err := connectDB(dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Warn("Failed to close database connection:", "error", err)
		}
	}()

	migrationsDir, err := validateMigrationsDir(*dir)
	if err != nil {
		return err
	}

	return executeCommand(db, migrationsDir, *command, *version)
}

func connectDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func validateMigrationsDir(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for migrations directory: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return "", fmt.Errorf("migrations directory does not exist: %s", absDir)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", absDir)
	}

	return absDir, nil
}

func executeCommand(db *sql.DB, migrationsDir, command string, version int64) error {
	switch command {
	case "up":
		return runUp(db, migrationsDir)
	case "down":
		return runDown(db, migrationsDir)
	case "down-to":
		return runDownTo(db, migrationsDir, version)
	case "up-to":
		return runUpTo(db, migrationsDir, version)
	case "status":
		return runStatus(db, migrationsDir)
	case "version":
		return runVersion(db)
	case "create":
		return runCreate(db, migrationsDir)
	default:
		return fmt.Errorf("unknown command: %s. Available commands: up, down, down-to, up-to, status, version, create", command)
	}
}

func runUp(db *sql.DB, migrationsDir string) error {
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations applied successfully")

	return nil
}

func runDown(db *sql.DB, migrationsDir string) error {
	if err := goose.Down(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	log.Println("Migrations rolled back successfully")

	return nil
}

func runDownTo(db *sql.DB, migrationsDir string, version int64) error {
	if version == 0 {
		return errors.New("version is required for down-to command")
	}

	if err := goose.DownTo(db, migrationsDir, version); err != nil {
		return fmt.Errorf("failed to rollback migrations to version %d: %w", version, err)
	}

	log.Printf("Migrations rolled back to version %d successfully", version)

	return nil
}

func runUpTo(db *sql.DB, migrationsDir string, version int64) error {
	if version == 0 {
		return errors.New("version is required for up-to command")
	}

	if err := goose.UpTo(db, migrationsDir, version); err != nil {
		return fmt.Errorf("failed to run migrations to version %d: %w", version, err)
	}

	log.Printf("Migrations applied to version %d successfully", version)

	return nil
}

func runStatus(db *sql.DB, migrationsDir string) error {
	if err := goose.Status(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}

func runVersion(db *sql.DB) error {
	version, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get database version: %w", err)
	}

	log.Printf("Current database version: %d", version)

	return nil
}

func runCreate(db *sql.DB, migrationsDir string) error {
	if len(flag.Args()) == 0 {
		return errors.New("migration name is required for create command")
	}

	name := flag.Args()[0]
	if err := goose.Create(db, migrationsDir, name, "sql"); err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	log.Printf("Migration created successfully: %s", name)

	return nil
}
