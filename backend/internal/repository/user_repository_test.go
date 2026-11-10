package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockRow struct {
	scanFunc func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.scanFunc != nil {
		return m.scanFunc(dest...)
	}
	return nil
}

type mockDBTX struct {
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	execFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	queryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func (m *mockDBTX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return &mockRow{}
}

func (m *mockDBTX) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, args...)
	}
	return pgconn.NewCommandTag(""), nil
}

func (m *mockDBTX) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func TestPostgresUserRepository_CreateUser(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	genID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*uuid.UUID) = genID
						*dest[1].(*time.Time) = now
						*dest[2].(*time.Time) = now
						return nil
					},
				}
			},
		}

		repo := NewUserRepository(mockDB)
		user := &models.User{
			Name:         "John Doe",
			Email:        "John.Doe@Example.com",
			PasswordHash: "hashed_secret",
		}

		err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if user.ID != genID {
			t.Errorf("expected ID %v, got %v", genID, user.ID)
		}
		if user.Email != "john.doe@example.com" {
			t.Errorf("expected lowercased email, got %s", user.Email)
		}
		if user.CreatedAt != now || user.UpdatedAt != now {
			t.Errorf("expected timestamps to be set to %v", now)
		}
	})

	t.Run("nil user", func(t *testing.T) {
		repo := NewUserRepository(&mockDBTX{})
		err := repo.CreateUser(ctx, nil)
		if !errors.Is(err, ErrNilUser) {
			t.Fatalf("expected ErrNilUser, got %v", err)
		}
	})

	t.Run("nil database", func(t *testing.T) {
		repo := NewUserRepository(nil)
		err := repo.CreateUser(ctx, &models.User{Name: "Alice"})
		if !errors.Is(err, ErrNilDatabase) {
			t.Fatalf("expected ErrNilDatabase, got %v", err)
		}
	})

	t.Run("duplicate email unique violation", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						return &pgconn.PgError{Code: "23505"}
					},
				}
			},
		}

		repo := NewUserRepository(mockDB)
		user := &models.User{
			Name:         "Existing",
			Email:        "existing@example.com",
			PasswordHash: "hashed_secret",
		}

		err := repo.CreateUser(ctx, user)
		if !errors.Is(err, ErrDuplicateEmail) {
			t.Fatalf("expected ErrDuplicateEmail, got %v", err)
		}
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						return errors.New("connection reset by peer")
					},
				}
			},
		}

		repo := NewUserRepository(mockDB)
		user := &models.User{Name: "Bob", Email: "bob@example.com"}

		err := repo.CreateUser(ctx, user)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestPostgresUserRepository_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()
	phone := "+1234567890"

	t.Run("success", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*uuid.UUID) = userID
						*dest[1].(*string) = "Alice Smith"
						*dest[2].(*string) = "alice@example.com"
						*dest[3].(*string) = "hashed_pw"
						*dest[4].(**string) = &phone
						*dest[5].(**string) = nil
						*dest[6].(*time.Time) = now
						*dest[7].(*time.Time) = now
						return nil
					},
				}
			},
		}

		repo := NewUserRepository(mockDB)
		user, err := repo.GetUserByEmail(ctx, " Alice@example.com ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if user.ID != userID {
			t.Errorf("expected user ID %v, got %v", userID, user.ID)
		}
		if user.Email != "alice@example.com" {
			t.Errorf("expected email alice@example.com, got %s", user.Email)
		}
		if user.Phone == nil || *user.Phone != phone {
			t.Errorf("expected phone %s, got %v", phone, user.Phone)
		}
		if user.AvatarURL != nil {
			t.Errorf("expected nil avatar, got %v", user.AvatarURL)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						return pgx.ErrNoRows
					},
				}
			},
		}

		repo := NewUserRepository(mockDB)
		user, err := repo.GetUserByEmail(ctx, "notfound@example.com")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
		if user != nil {
			t.Fatalf("expected nil user, got %v", user)
		}
	})

	t.Run("empty email", func(t *testing.T) {
		repo := NewUserRepository(&mockDBTX{})
		user, err := repo.GetUserByEmail(ctx, "   ")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound for empty email, got %v", err)
		}
		if user != nil {
			t.Fatalf("expected nil user, got %v", user)
		}
	})

	t.Run("nil database", func(t *testing.T) {
		repo := NewUserRepository(nil)
		_, err := repo.GetUserByEmail(ctx, "test@example.com")
		if !errors.Is(err, ErrNilDatabase) {
			t.Fatalf("expected ErrNilDatabase, got %v", err)
		}
	})
}

func TestPostgresUserRepository_GetUserByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*uuid.UUID) = userID
						*dest[1].(*string) = "Bob"
						*dest[2].(*string) = "bob@example.com"
						*dest[3].(*string) = "hashed_secret"
						*dest[4].(**string) = nil
						*dest[5].(**string) = nil
						*dest[6].(*time.Time) = now
						*dest[7].(*time.Time) = now
						return nil
					},
				}
			},
		}

		repo := NewUserRepository(mockDB)
		user, err := repo.GetUserByID(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if user.ID != userID {
			t.Errorf("expected user ID %v, got %v", userID, user.ID)
		}
		if user.Name != "Bob" {
			t.Errorf("expected name Bob, got %s", user.Name)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						return pgx.ErrNoRows
					},
				}
			},
		}

		repo := NewUserRepository(mockDB)
		user, err := repo.GetUserByID(ctx, userID)
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
		if user != nil {
			t.Fatalf("expected nil user, got %v", user)
		}
	})

	t.Run("nil uuid", func(t *testing.T) {
		repo := NewUserRepository(&mockDBTX{})
		user, err := repo.GetUserByID(ctx, uuid.Nil)
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound for nil UUID, got %v", err)
		}
		if user != nil {
			t.Fatalf("expected nil user, got %v", user)
		}
	})

	t.Run("nil database", func(t *testing.T) {
		repo := NewUserRepository(nil)
		_, err := repo.GetUserByID(ctx, userID)
		if !errors.Is(err, ErrNilDatabase) {
			t.Fatalf("expected ErrNilDatabase, got %v", err)
		}
	})
}
