package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrUserNotFound is returned when a requested user record cannot be found.
	ErrUserNotFound = errors.New("user not found")
	// ErrDuplicateEmail is returned when an operation violates the unique constraint on email.
	ErrDuplicateEmail = errors.New("email is already registered")
	// ErrNilUser is returned when a nil user pointer is provided.
	ErrNilUser = errors.New("user cannot be nil")
	// ErrNilDatabase is returned when database connection is nil.
	ErrNilDatabase = errors.New("database connection is not initialized")
)

// DBTX defines the common query execution methods shared by *pgxpool.Pool and pgx.Tx.
type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// UserRepository defines the persistence interface for user records.
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// PostgresUserRepository implements UserRepository using a PostgreSQL connection pool or transaction.
type PostgresUserRepository struct {
	db DBTX
}

// NewUserRepository constructs a new PostgresUserRepository.
func NewUserRepository(db DBTX) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// CreateUser inserts a new user record into PostgreSQL and scans generated fields back into the model.
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	if r.db == nil {
		return ErrNilDatabase
	}
	if user == nil {
		return ErrNilUser
	}

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	query := `
		INSERT INTO users (id, name, email, password_hash, phone, avatar_url)
		VALUES (COALESCE(NULLIF($1, '00000000-0000-0000-0000-000000000000'::uuid), gen_random_uuid()), $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Phone,
		user.AvatarURL,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByEmail queries a user record by email address (case-insensitive).
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if r.db == nil {
		return nil, ErrNilDatabase
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if cleanEmail == "" {
		return nil, ErrUserNotFound
	}

	query := `
		SELECT id, name, email, password_hash, phone, avatar_url, created_at, updated_at
		FROM users
		WHERE LOWER(email) = $1
		LIMIT 1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, cleanEmail).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Phone,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetUserByID queries a user record by its primary key UUID.
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if r.db == nil {
		return nil, ErrNilDatabase
	}
	if id == uuid.Nil {
		return nil, ErrUserNotFound
	}

	query := `
		SELECT id, name, email, password_hash, phone, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Phone,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}
