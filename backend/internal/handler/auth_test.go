package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/MohammedaminSa/FindIt/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockAuthService struct {
	registerFunc       func(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error)
	loginFunc          func(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error)
	getUserProfileFunc func(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error)
}

func (m *mockAuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockAuthService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error) {
	if m.getUserProfileFunc != nil {
		return m.getUserProfileFunc(ctx, userID)
	}
	return nil, nil
}

func setupAuthTestRouter(h *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	h.RegisterRoutes(v1)
	return router
}

func TestAuthHandler_Register(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		mockSvc := &mockAuthService{
			registerFunc: func(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
				return &models.AuthResponse{
					Token: "test-token-123",
					User: &models.UserResponse{
						ID:        userID,
						Name:      req.Name,
						Email:     req.Email,
						CreatedAt: now,
						UpdatedAt: now,
					},
				}, nil
			},
		}

		h := NewAuthHandler(mockSvc)
		router := setupAuthTestRouter(h)

		payload := `{"name":"Alice","email":"alice@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected HTTP 201, got %d: %s", w.Code, w.Body.String())
		}

		var resp models.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Token != "test-token-123" {
			t.Errorf("expected token 'test-token-123', got %s", resp.Token)
		}
		if resp.User.Email != "alice@example.com" {
			t.Errorf("expected email 'alice@example.com', got %s", resp.User.Email)
		}
	})

	t.Run("validation error - missing fields", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthService{})
		router := setupAuthTestRouter(h)

		payload := `{"email":"not-an-email"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected HTTP 400, got %d", w.Code)
		}
	})

	t.Run("duplicate email conflict", func(t *testing.T) {
		mockSvc := &mockAuthService{
			registerFunc: func(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
				return nil, service.ErrEmailAlreadyInUse
			},
		}

		h := NewAuthHandler(mockSvc)
		router := setupAuthTestRouter(h)

		payload := `{"name":"Alice","email":"alice@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected HTTP 409, got %d", w.Code)
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		mockSvc := &mockAuthService{
			registerFunc: func(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
				return nil, errors.New("db error")
			},
		}

		h := NewAuthHandler(mockSvc)
		router := setupAuthTestRouter(h)

		payload := `{"name":"Alice","email":"alice@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected HTTP 500, got %d", w.Code)
		}
	})
}

func TestAuthHandler_Login(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := &mockAuthService{
			loginFunc: func(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
				return &models.AuthResponse{
					Token: "login-jwt-token",
					User: &models.UserResponse{
						ID:    userID,
						Email: req.Email,
					},
				}, nil
			},
		}

		h := NewAuthHandler(mockSvc)
		router := setupAuthTestRouter(h)

		payload := `{"email":"bob@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d", w.Code)
		}

		var resp models.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if resp.Token != "login-jwt-token" {
			t.Errorf("expected token 'login-jwt-token', got %s", resp.Token)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthService{})
		router := setupAuthTestRouter(h)

		payload := `{"email":"invalid-email"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected HTTP 400, got %d", w.Code)
		}
	})

	t.Run("invalid credentials unauthorized", func(t *testing.T) {
		mockSvc := &mockAuthService{
			loginFunc: func(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
				return nil, service.ErrInvalidCredentials
			},
		}

		h := NewAuthHandler(mockSvc)
		router := setupAuthTestRouter(h)

		payload := `{"email":"bob@example.com","password":"wrongpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected HTTP 401, got %d", w.Code)
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		mockSvc := &mockAuthService{
			loginFunc: func(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
				return nil, errors.New("database timeout")
			},
		}

		h := NewAuthHandler(mockSvc)
		router := setupAuthTestRouter(h)

		payload := `{"email":"bob@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected HTTP 500, got %d", w.Code)
		}
	})
}

func TestAuthHandler_Me(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockSvc := &mockAuthService{
			getUserProfileFunc: func(ctx context.Context, id uuid.UUID) (*models.UserResponse, error) {
				return &models.UserResponse{
					ID:    id,
					Name:  "Charlie",
					Email: "charlie@example.com",
				}, nil
			},
		}

		h := NewAuthHandler(mockSvc)
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/me", func(c *gin.Context) {
			c.Set("userID", userID)
			h.Me(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d", w.Code)
		}
	})

	t.Run("unauthorized - no user in context", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthService{})
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/me", h.Me)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected HTTP 401, got %d", w.Code)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockSvc := &mockAuthService{
			getUserProfileFunc: func(ctx context.Context, id uuid.UUID) (*models.UserResponse, error) {
				return nil, service.ErrUserNotFound
			},
		}

		h := NewAuthHandler(mockSvc)
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/me", func(c *gin.Context) {
			c.Set("userID", userID)
			h.Me(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected HTTP 404, got %d", w.Code)
		}
	})
}
