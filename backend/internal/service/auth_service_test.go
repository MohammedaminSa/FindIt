package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/MohammedaminSa/FindIt/backend/internal/repository"
	"github.com/MohammedaminSa/FindIt/backend/pkg/hasher"
	"github.com/MohammedaminSa/FindIt/backend/pkg/token"
	"github.com/google/uuid"
)

type mockUserRepo struct {
	createUserFunc     func(ctx context.Context, user *models.User) error
	getUserByEmailFunc func(ctx context.Context, email string) (*models.User, error)
	getUserByIDFunc    func(ctx context.Context, id uuid.UUID) (*models.User, error)
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	return nil, repository.ErrUserNotFound
}

type mockTokenManager struct {
	generateTokenFunc func(userID uuid.UUID, email string) (string, error)
	validateTokenFunc func(tokenString string) (*token.Claims, error)
}

func (m *mockTokenManager) GenerateToken(userID uuid.UUID, email string) (string, error) {
	if m.generateTokenFunc != nil {
		return m.generateTokenFunc(userID, email)
	}
	return "mock-jwt-token-xyz", nil
}

func (m *mockTokenManager) ValidateToken(tokenString string) (*token.Claims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(tokenString)
	}
	return &token.Claims{}, nil
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		genID := uuid.New()
		now := time.Now().UTC()

		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			createUserFunc: func(ctx context.Context, user *models.User) error {
				user.ID = genID
				user.CreatedAt = now
				user.UpdatedAt = now
				return nil
			},
		}

		tokenMgr := &mockTokenManager{
			generateTokenFunc: func(userID uuid.UUID, email string) (string, error) {
				return "generated-jwt-token", nil
			},
		}

		svc := NewAuthService(repo, tokenMgr)
		req := &models.RegisterRequest{
			Name:     "Alice Smith",
			Email:    "Alice@Example.COM",
			Password: "securepassword123",
			Phone:    "+1234567890",
		}

		resp, err := svc.Register(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Token != "generated-jwt-token" {
			t.Errorf("expected token 'generated-jwt-token', got %s", resp.Token)
		}
		if resp.User.ID != genID {
			t.Errorf("expected user ID %v, got %v", genID, resp.User.ID)
		}
		if resp.User.Email != "alice@example.com" {
			t.Errorf("expected normalized email alice@example.com, got %s", resp.User.Email)
		}
		if resp.User.Phone == nil || *resp.User.Phone != "+1234567890" {
			t.Errorf("expected phone '+1234567890', got %v", resp.User.Phone)
		}
	})

	t.Run("nil request", func(t *testing.T) {
		svc := NewAuthService(&mockUserRepo{}, &mockTokenManager{})
		_, err := svc.Register(ctx, nil)
		if !errors.Is(err, ErrNilRequest) {
			t.Fatalf("expected ErrNilRequest, got %v", err)
		}
	})

	t.Run("empty email", func(t *testing.T) {
		svc := NewAuthService(&mockUserRepo{}, &mockTokenManager{})
		_, err := svc.Register(ctx, &models.RegisterRequest{
			Name:     "Alice",
			Email:    "   ",
			Password: "password123",
		})
		if err == nil {
			t.Fatal("expected error for empty email, got nil")
		}
	})

	t.Run("email already in use proactively", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &models.User{ID: uuid.New(), Email: email}, nil
			},
		}

		svc := NewAuthService(repo, &mockTokenManager{})
		_, err := svc.Register(ctx, &models.RegisterRequest{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: "password123",
		})
		if !errors.Is(err, ErrEmailAlreadyInUse) {
			t.Fatalf("expected ErrEmailAlreadyInUse, got %v", err)
		}
	})

	t.Run("duplicate email caught on create", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			createUserFunc: func(ctx context.Context, user *models.User) error {
				return repository.ErrDuplicateEmail
			},
		}

		svc := NewAuthService(repo, &mockTokenManager{})
		_, err := svc.Register(ctx, &models.RegisterRequest{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: "password123",
		})
		if !errors.Is(err, ErrEmailAlreadyInUse) {
			t.Fatalf("expected ErrEmailAlreadyInUse from CreateUser, got %v", err)
		}
	})

	t.Run("password too long (hasher error)", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
		}

		svc := NewAuthService(repo, &mockTokenManager{})
		_, err := svc.Register(ctx, &models.RegisterRequest{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: strings.Repeat("A", 73),
		})
		if err == nil {
			t.Fatal("expected error for password exceeding 72 bytes, got nil")
		}
	})

	t.Run("token manager error", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			createUserFunc: func(ctx context.Context, user *models.User) error {
				user.ID = uuid.New()
				return nil
			},
		}
		tokenMgr := &mockTokenManager{
			generateTokenFunc: func(userID uuid.UUID, email string) (string, error) {
				return "", errors.New("jwt signing failed")
			},
		}

		svc := NewAuthService(repo, tokenMgr)
		_, err := svc.Register(ctx, &models.RegisterRequest{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: "password123",
		})
		if err == nil {
			t.Fatal("expected token generation error, got nil")
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		plainPassword := "secretpassword123"
		pwHash, _ := hasher.HashPassword(plainPassword)
		userID := uuid.New()

		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &models.User{
					ID:           userID,
					Name:         "Bob",
					Email:        "bob@example.com",
					PasswordHash: pwHash,
				}, nil
			},
		}

		tokenMgr := &mockTokenManager{
			generateTokenFunc: func(uID uuid.UUID, email string) (string, error) {
				return "bob-jwt-token", nil
			},
		}

		svc := NewAuthService(repo, tokenMgr)
		resp, err := svc.Login(ctx, &models.LoginRequest{
			Email:    "BOB@example.com",
			Password: plainPassword,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Token != "bob-jwt-token" {
			t.Errorf("expected token bob-jwt-token, got %s", resp.Token)
		}
		if resp.User.ID != userID {
			t.Errorf("expected user ID %v, got %v", userID, resp.User.ID)
		}
	})

	t.Run("nil request", func(t *testing.T) {
		svc := NewAuthService(&mockUserRepo{}, &mockTokenManager{})
		_, err := svc.Login(ctx, nil)
		if !errors.Is(err, ErrNilRequest) {
			t.Fatalf("expected ErrNilRequest, got %v", err)
		}
	})

	t.Run("empty email", func(t *testing.T) {
		svc := NewAuthService(&mockUserRepo{}, &mockTokenManager{})
		_, err := svc.Login(ctx, &models.LoginRequest{
			Email:    "   ",
			Password: "pass",
		})
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
		}

		svc := NewAuthService(repo, &mockTokenManager{})
		_, err := svc.Login(ctx, &models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		})
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		pwHash, _ := hasher.HashPassword("correctpassword")

		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &models.User{
					ID:           uuid.New(),
					Email:        email,
					PasswordHash: pwHash,
				}, nil
			},
		}

		svc := NewAuthService(repo, &mockTokenManager{})
		_, err := svc.Login(ctx, &models.LoginRequest{
			Email:    "user@example.com",
			Password: "wrongpassword",
		})
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
		}
	})

	t.Run("database error on get user", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, errors.New("db connection failure")
			},
		}

		svc := NewAuthService(repo, &mockTokenManager{})
		_, err := svc.Login(ctx, &models.LoginRequest{
			Email:    "user@example.com",
			Password: "password123",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAuthService_GetUserProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:    userID,
					Name:  "Charlie",
					Email: "charlie@example.com",
				}, nil
			},
		}

		svc := NewAuthService(repo, &mockTokenManager{})
		profile, err := svc.GetUserProfile(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if profile.ID != userID {
			t.Errorf("expected user ID %v, got %v", userID, profile.ID)
		}
		if profile.Name != "Charlie" {
			t.Errorf("expected name Charlie, got %s", profile.Name)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
		}

		svc := NewAuthService(repo, &mockTokenManager{})
		_, err := svc.GetUserProfile(ctx, userID)
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("database error", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, errors.New("unexpected error")
			},
		}

		svc := NewAuthService(repo, &mockTokenManager{})
		_, err := svc.GetUserProfile(ctx, userID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
