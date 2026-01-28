package repository

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	rpcv1 "github.com/notification-system-moxicom/persistence-service/pkg/proto/gen/persistence/v1"
)

type SystemRepository interface {
	CreateSystem(ctx context.Context, name, description string) (*rpcv1.System, error)
	ListSystems(ctx context.Context) ([]*rpcv1.System, error)
}

type UserRepository interface {
	AddUser(ctx context.Context, systemID, idAtSystem string, adapters *rpcv1.Adapter) (*rpcv1.User, error)
	ListUsers(ctx context.Context, systemID string) ([]*rpcv1.User, error)
}

type Repository interface {
	SystemRepository
	UserRepository
}

type postgresRep struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
}

// NewpostgresRep constructs a new Postgres-backed repository.
func New(pool *pgxpool.Pool) Repository {
	return &postgresRep{
		pool: pool,
		sb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// CreateSystem inserts a new system and returns it.
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

// ListSystems returns all systems.
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

// AddUser inserts a new user and returns it.
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

// ListUsers returns all users for a given system.
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
