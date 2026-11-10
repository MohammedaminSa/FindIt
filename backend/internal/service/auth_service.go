package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/MohammedaminSa/FindIt/backend/internal/repository"
	"github.com/MohammedaminSa/FindIt/backend/pkg/hasher"
	"github.com/MohammedaminSa/FindIt/backend/pkg/token"
	"github.com/google/uuid"
)

var (
	// ErrInvalidCredentials is returned when email or password verification fails.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrEmailAlreadyInUse is returned when attempting to register with an email that is already registered.
	ErrEmailAlreadyInUse = errors.New("email is already registered")
	// ErrUserNotFound is returned when a requested user profile does not exist.
	ErrUserNotFound = errors.New("user not found")
	// ErrNilRequest is returned when an operation is passed a nil request payload.
	ErrNilRequest = errors.New("request payload cannot be nil")
)

// AuthService defines business operations for user authentication and accounts.
type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error)
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error)
}

// DefaultAuthService implements AuthService.
type DefaultAuthService struct {
	userRepo     repository.UserRepository
	tokenManager token.TokenManager
}

// NewAuthService constructs a new DefaultAuthService with its dependencies.
func NewAuthService(userRepo repository.UserRepository, tokenManager token.TokenManager) *DefaultAuthService {
	return &DefaultAuthService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

// Register validates user registration data, hashes the password, creates the user record, and generates a JWT.
func (s *DefaultAuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	if cleanEmail == "" {
		return nil, errors.New("email cannot be empty")
	}

	// 1. Proactively check if email is already registered
	existingUser, err := s.userRepo.GetUserByEmail(ctx, cleanEmail)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyInUse
	}

	// 2. Hash user password securely
	passwordHash, err := hasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. Construct user model
	user := &models.User{
		Name:         strings.TrimSpace(req.Name),
		Email:        cleanEmail,
		PasswordHash: passwordHash,
	}
	if req.Phone != "" {
		trimmedPhone := strings.TrimSpace(req.Phone)
		user.Phone = &trimmedPhone
	}

	// 4. Persist user to database
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, ErrEmailAlreadyInUse
		}
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// 5. Generate signed JWT token
	tokenStr, err := s.tokenManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	return &models.AuthResponse{
		Token: tokenStr,
		User:  user.ToResponse(),
	}, nil
}

// Login verifies user credentials against stored bcrypt hash and returns a signed JWT token.
func (s *DefaultAuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	if cleanEmail == "" {
		return nil, ErrInvalidCredentials
	}

	// 1. Fetch user by email
	user, err := s.userRepo.GetUserByEmail(ctx, cleanEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// Intentionally return generic error to prevent account enumeration
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Compare password hash
	if !hasher.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// 3. Generate signed JWT token
	tokenStr, err := s.tokenManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	return &models.AuthResponse{
		Token: tokenStr,
		User:  user.ToResponse(),
	}, nil
}

// GetUserProfile retrieves the safe user profile by ID.
func (s *DefaultAuthService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return user.ToResponse(), nil
}
