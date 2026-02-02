package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	rpcv1 "github.com/notification-system-moxicom/persistence-service/pkg/proto/gen/persistence/v1"
)

type Config struct {
	UserNameENV string `yaml:"username_env"`
	PassENV     string `yaml:"password_env"`
	HostENV     string `yaml:"host_env"`
	DSNWithEnv  string `yaml:"dsn_with_env"` // example: postgres://{{DB_USERNAME}}:{{DB_PASSWORD}}@{{DB_HOST}}:5432/database_name
}

type SystemRepository interface {
	CreateSystem(ctx context.Context, name, description string) (*rpcv1.System, error)
	ListSystems(ctx context.Context) ([]*rpcv1.System, error)
	UpdateSystem(ctx context.Context, id string, name, description *string) (*rpcv1.System, error)
	DeleteSystem(ctx context.Context, id string) error
}

type UserRepository interface {
	AddUser(ctx context.Context, systemID, idAtSystem string, adapters *rpcv1.Adapter) (*rpcv1.User, error)
	ListUsers(ctx context.Context, systemID string) ([]*rpcv1.User, error)
	UpdateUser(ctx context.Context, id string, idAtSystem *string, adapters *rpcv1.Adapter) (*rpcv1.User, error)
	DeleteUser(ctx context.Context, id string) error
}

type Repository interface {
	SystemRepository
	UserRepository
}

type postgresRep struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
}

// New constructs a new Postgres-backed repository.
func New(ctx context.Context, cfg Config) (Repository, error) {
	dsn := BuildDSN(cfg)
	if dsn == "" {
		return nil, errors.New("failed to build DSN from config")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("failed to create postgres pool:", slog.String("error", err.Error()))
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &postgresRep{
		pool: pool,
		sb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func BuildDSN(cfg Config) string {
	if cfg.DSNWithEnv != "" {
		dsn := cfg.DSNWithEnv
		if cfg.UserNameENV != "" {
			envValue := os.Getenv(cfg.UserNameENV)
			dsn = strings.ReplaceAll(dsn, "{{DB_USERNAME}}", envValue)
		}

		if cfg.PassENV != "" {
			envValue := os.Getenv(cfg.PassENV)
			dsn = strings.ReplaceAll(dsn, "{{DB_PASSWORD}}", envValue)
		}

		if cfg.HostENV != "" {
			envValue := os.Getenv(cfg.HostENV)
			dsn = strings.ReplaceAll(dsn, "{{DB_HOST}}", envValue)
		}

		return dsn
	}

	return ""
}

func (r *postgresRep) CreateSystem(
	ctx context.Context,
	name,
	description string,
) (*rpcv1.System, error) {
	now := time.Now().UTC()

	query := r.sb.
		Insert("systems").
		Columns("name", "description", "created_at", "updated_at").
		Values(name, description, now, now).
		Suffix("RETURNING id, name, description, created_at, updated_at, deleted_at")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, sqlStr, args...)

	var (
		id        string
		createdAt time.Time
		updatedAt time.Time
		deletedAt sql.NullTime
	)

	if err := row.Scan(&id, &name, &description, &createdAt, &updatedAt, &deletedAt); err != nil {
		return nil, err
	}

	system := &rpcv1.System{
		Id:          id,
		Name:        name,
		Description: description,
		CreatedAt:   createdAt.Unix(),
		UpdatedAt:   updatedAt.Unix(),
	}

	if deletedAt.Valid {
		system.DeletedAt = deletedAt.Time.Unix()
	}

	return system, nil
}

func (r *postgresRep) ListSystems(ctx context.Context) ([]*rpcv1.System, error) {
	query := r.sb.
		Select("id", "name", "description", "created_at", "updated_at", "deleted_at").
		From("systems")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var systems []*rpcv1.System

	for rows.Next() {
		var (
			id          string
			name        string
			description string
			createdAt   time.Time
			updatedAt   time.Time
			deletedAt   sql.NullTime
		)

		if err := rows.Scan(&id, &name, &description, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}

		system := &rpcv1.System{
			Id:          id,
			Name:        name,
			Description: description,
			CreatedAt:   createdAt.Unix(),
			UpdatedAt:   updatedAt.Unix(),
		}

		if deletedAt.Valid {
			system.DeletedAt = deletedAt.Time.Unix()
		}

		systems = append(systems, system)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return systems, nil
}

func (r *postgresRep) AddUser(
	ctx context.Context,
	systemID,
	idAtSystem string,
	adapters *rpcv1.Adapter,
) (*rpcv1.User, error) {
	query := r.sb.
		Insert("users").
		Columns("system_id", "id_at_system", "email", "phone", "telegram_chat_id").
		Values(systemID, idAtSystem, adapters.GetEmail(), adapters.GetPhone(), adapters.GetTelegramChatId()).
		Suffix("RETURNING id, system_id, id_at_system, email, phone, telegram_chat_id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, sqlStr, args...)

	var (
		id            string
		outSystemID   string
		outIDAtSystem string
		email         string
		phone         string
		telegramID    string
	)

	if err := row.Scan(&id, &outSystemID, &outIDAtSystem, &email, &phone, &telegramID); err != nil {
		return nil, err
	}

	return &rpcv1.User{
		Id:         id,
		IdAtSystem: outIDAtSystem,
		Adapters: &rpcv1.Adapter{
			Email:          email,
			Phone:          phone,
			TelegramChatId: telegramID,
		},
	}, nil
}

func (r *postgresRep) ListUsers(ctx context.Context, systemID string) ([]*rpcv1.User, error) {
	query := r.sb.
		Select("id", "system_id", "id_at_system", "email", "phone", "telegram_chat_id").
		From("users").
		Where(sq.Eq{"system_id": systemID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*rpcv1.User

	for rows.Next() {
		var (
			id            string
			outSystemID   string
			outIDAtSystem string
			email         string
			phone         string
			telegramID    string
		)

		if err := rows.Scan(&id, &outSystemID, &outIDAtSystem, &email, &phone, &telegramID); err != nil {
			return nil, err
		}

		users = append(users, &rpcv1.User{
			Id:         id,
			IdAtSystem: outIDAtSystem,
			Adapters: &rpcv1.Adapter{
				Email:          email,
				Phone:          phone,
				TelegramChatId: telegramID,
			},
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *postgresRep) UpdateSystem(
	ctx context.Context,
	id string,
	name, description *string,
) (*rpcv1.System, error) {
	now := time.Now().UTC()
	query := r.sb.Update("systems").Set("updated_at", now)

	if name != nil {
		query = query.Set("name", *name)
	}

	if description != nil {
		query = query.Set("description", *description)
	}

	query = query.Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, name, description, created_at, updated_at, deleted_at")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, sqlStr, args...)

	var (
		outID     string
		outName   string
		outDesc   string
		createdAt time.Time
		updatedAt time.Time
		deletedAt sql.NullTime
	)

	if err := row.Scan(&outID, &outName, &outDesc, &createdAt, &updatedAt, &deletedAt); err != nil {
		return nil, err
	}

	system := &rpcv1.System{
		Id:          outID,
		Name:        outName,
		Description: outDesc,
		CreatedAt:   createdAt.Unix(),
		UpdatedAt:   updatedAt.Unix(),
	}

	if deletedAt.Valid {
		system.DeletedAt = deletedAt.Time.Unix()
	}

	return system, nil
}

func (r *postgresRep) DeleteSystem(ctx context.Context, id string) error {
	now := time.Now().UTC()
	query := r.sb.
		Update("systems").
		Set("deleted_at", now).
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sqlStr, args...)

	return err
}

func (r *postgresRep) UpdateUser(
	ctx context.Context,
	id string,
	idAtSystem *string,
	adapters *rpcv1.Adapter,
) (*rpcv1.User, error) {
	query := r.sb.Update("users")

	if idAtSystem != nil {
		query = query.Set("id_at_system", *idAtSystem)
	}

	if adapters != nil {
		query = query.Set("email", adapters.GetEmail())
		query = query.Set("phone", adapters.GetPhone())
		query = query.Set("telegram_chat_id", adapters.GetTelegramChatId())
	}

	query = query.Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, system_id, id_at_system, email, phone, telegram_chat_id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, sqlStr, args...)

	var (
		outID         string
		outSystemID   string
		outIDAtSystem string
		email         string
		phone         string
		telegramID    string
	)

	if err := row.Scan(&outID, &outSystemID, &outIDAtSystem, &email, &phone, &telegramID); err != nil {
		return nil, err
	}

	return &rpcv1.User{
		Id:         outID,
		IdAtSystem: outIDAtSystem,
		Adapters: &rpcv1.Adapter{
			Email:          email,
			Phone:          phone,
			TelegramChatId: telegramID,
		},
	}, nil
}

func (r *postgresRep) DeleteUser(ctx context.Context, id string) error {
	query := r.sb.
		Delete("users").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sqlStr, args...)

	return err
}
