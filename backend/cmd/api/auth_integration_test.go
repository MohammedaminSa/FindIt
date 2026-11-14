package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/handler"
	"github.com/MohammedaminSa/FindIt/backend/internal/middleware"
	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/MohammedaminSa/FindIt/backend/internal/repository"
	"github.com/MohammedaminSa/FindIt/backend/internal/service"
	"github.com/MohammedaminSa/FindIt/backend/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// inMemoryUserRepo provides a thread-safe in-memory store for integration testing.
type inMemoryUserRepo struct {
	mu          sync.RWMutex
	usersByID   map[uuid.UUID]*models.User
	usersByMail map[string]*models.User
}

func newInMemoryUserRepo() *inMemoryUserRepo {
	return &inMemoryUserRepo{
		usersByID:   make(map[uuid.UUID]*models.User),
		usersByMail: make(map[string]*models.User),
	}
}

func (r *inMemoryUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user == nil {
		return repository.ErrNilUser
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))
	if _, exists := r.usersByMail[normalizedEmail]; exists {
		return repository.ErrDuplicateEmail
	}

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	cloned := *user
	r.usersByID[user.ID] = &cloned
	r.usersByMail[normalizedEmail] = &cloned
	return nil
}

func (r *inMemoryUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalized := strings.ToLower(strings.TrimSpace(email))
	u, exists := r.usersByMail[normalized]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	cloned := *u
	return &cloned, nil
}

func (r *inMemoryUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.usersByID[id]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	cloned := *u
	return &cloned, nil
}

func setupIntegrationServer(t *testing.T) (*gin.Engine, *inMemoryUserRepo, token.TokenManager) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	repo := newInMemoryUserRepo()
	jwtMgr, err := token.NewJWTManager("integration-test-secret-key-at-least-32-chars!", time.Hour)
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}

	authService := service.NewAuthService(repo, jwtMgr)
	authHandler := handler.NewAuthHandler(authService)
	authMw := middleware.Auth(jwtMgr)

	router := setupRouter(nil, authHandler, authMw)
	return router, repo, jwtMgr
}

// TestAuth_EndToEndLifecycle verifies the complete auth journey:
// 1. Register a user -> 201 Created with JWT & User profile
// 2. Login with credentials -> 200 OK with fresh JWT
// 3. Access GET /api/v1/auth/me with Bearer token -> 200 OK with user profile
// 4. Duplicate registration -> 409 Conflict
// 5. Invalid password login -> 401 Unauthorized
// 6. Unknown email login -> 401 Unauthorized
func TestAuth_EndToEndLifecycle(t *testing.T) {
	router, _, _ := setupIntegrationServer(t)

	phone := "+1234567890"
	registerPayload := models.RegisterRequest{
		Name:     "Alice Walker",
		Email:    "alice.walker@example.com",
		Password: "SecurePassword123!",
		Phone:    phone,
	}

	// 1. Registration
	regBody, _ := json.Marshal(registerPayload)
	reqReg := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	router.ServeHTTP(wReg, reqReg)

	if wReg.Code != http.StatusCreated {
		t.Fatalf("expected registration HTTP 201, got %d: %s", wReg.Code, wReg.Body.String())
	}

	var regResponse models.AuthResponse
	if err := json.Unmarshal(wReg.Body.Bytes(), &regResponse); err != nil {
		t.Fatalf("failed to parse registration response: %v", err)
	}

	if regResponse.Token == "" {
		t.Error("expected non-empty JWT token on registration")
	}
	if regResponse.User == nil || regResponse.User.Email != registerPayload.Email {
		t.Errorf("expected user email %s, got %v", registerPayload.Email, regResponse.User)
	}
	if regResponse.User.Name != registerPayload.Name {
		t.Errorf("expected user name %s, got %s", registerPayload.Name, regResponse.User.Name)
	}
	if regResponse.User.Phone == nil || *regResponse.User.Phone != phone {
		t.Errorf("expected user phone %s, got %v", phone, regResponse.User.Phone)
	}

	// 2. Login
	loginPayload := models.LoginRequest{
		Email:    "Alice.Walker@example.com", // Case insensitivity check
		Password: "SecurePassword123!",
	}
	loginBody, _ := json.Marshal(loginPayload)
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected login HTTP 200, got %d: %s", wLogin.Code, wLogin.Body.String())
	}

	var loginResponse models.AuthResponse
	if err := json.Unmarshal(wLogin.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}

	if loginResponse.Token == "" {
		t.Error("expected non-empty JWT token on login")
	}

	// 3. Access Protected /api/v1/auth/me with Bearer token
	reqMe := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+loginResponse.Token)
	wMe := httptest.NewRecorder()
	router.ServeHTTP(wMe, reqMe)

	if wMe.Code != http.StatusOK {
		t.Fatalf("expected /auth/me HTTP 200, got %d: %s", wMe.Code, wMe.Body.String())
	}

	var meResponse models.UserResponse
	if err := json.Unmarshal(wMe.Body.Bytes(), &meResponse); err != nil {
		t.Fatalf("failed to parse /auth/me response: %v", err)
	}

	if meResponse.ID != regResponse.User.ID {
		t.Errorf("expected user ID %s, got %s", regResponse.User.ID, meResponse.ID)
	}
	if meResponse.Email != registerPayload.Email {
		t.Errorf("expected user email %s, got %s", registerPayload.Email, meResponse.Email)
	}

	// 4. Duplicate Registration Conflict
	reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	reqDup.Header.Set("Content-Type", "application/json")
	wDup := httptest.NewRecorder()
	router.ServeHTTP(wDup, reqDup)

	if wDup.Code != http.StatusConflict {
		t.Fatalf("expected duplicate registration HTTP 409, got %d", wDup.Code)
	}

	// 5. Login with Wrong Password
	badPassPayload := models.LoginRequest{
		Email:    registerPayload.Email,
		Password: "IncorrectPassword!",
	}
	badPassBody, _ := json.Marshal(badPassPayload)
	reqBadPass := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badPassBody))
	reqBadPass.Header.Set("Content-Type", "application/json")
	wBadPass := httptest.NewRecorder()
	router.ServeHTTP(wBadPass, reqBadPass)

	if wBadPass.Code != http.StatusUnauthorized {
		t.Fatalf("expected invalid password HTTP 401, got %d", wBadPass.Code)
	}

	// 6. Login with Non-Existent Email
	nonExistentPayload := models.LoginRequest{
		Email:    "unknown.user@example.com",
		Password: "SecurePassword123!",
	}
	nonExistentBody, _ := json.Marshal(nonExistentPayload)
	reqNonExistent := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(nonExistentBody))
	reqNonExistent.Header.Set("Content-Type", "application/json")
	wNonExistent := httptest.NewRecorder()
	router.ServeHTTP(wNonExistent, reqNonExistent)

	if wNonExistent.Code != http.StatusUnauthorized {
		t.Fatalf("expected unknown user HTTP 401, got %d", wNonExistent.Code)
	}
}

// TestAuth_ProtectedEndpointGuards validates that /api/v1/auth/me rejects unauthorized requests.
func TestAuth_ProtectedEndpointGuards(t *testing.T) {
	router, _, _ := setupIntegrationServer(t)

	foreignMgr, err := token.NewJWTManager("different-foreign-secret-key-at-least-32-chars!", time.Hour)
	if err != nil {
		t.Fatalf("failed to create foreign token manager: %v", err)
	}
	foreignToken, _ := foreignMgr.GenerateToken(uuid.New(), "foreign@example.com")

	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "missing authorization header",
			authHeader: "",
		},
		{
			name:       "unsupported basic auth scheme",
			authHeader: "Basic dXNlcjpwYXNz",
		},
		{
			name:       "malformed header without token",
			authHeader: "Bearer",
		},
		{
			name:       "malformed header with extra fields",
			authHeader: "Bearer token extra-part",
		},
		{
			name:       "tampered or invalid jwt token string",
			authHeader: "Bearer invalid.jwt.signature",
		},
		{
			name:       "token signed by a different secret key",
			authHeader: "Bearer " + foreignToken,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected HTTP 401 Unauthorized, got %d", w.Code)
			}
		})
	}

	t.Run("expired token rejection", func(t *testing.T) {
		expiredMgr, err := token.NewJWTManager("integration-test-secret-key-at-least-32-chars!", -1*time.Minute)
		if err != nil {
			t.Fatalf("failed to create expired token manager: %v", err)
		}
		expiredToken, err := expiredMgr.GenerateToken(uuid.New(), "expired@example.com")
		if err != nil {
			t.Fatalf("failed to generate expired token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected HTTP 401 for expired token, got %d", w.Code)
		}

		var resp map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["error"] != "token has expired" {
			t.Errorf("expected error message 'token has expired', got %q", resp["error"])
		}
	})
}

// TestAuth_InputValidation ensures invalid registration and login payloads are rejected with HTTP 400.
func TestAuth_InputValidation(t *testing.T) {
	router, _, _ := setupIntegrationServer(t)

	tests := []struct {
		name       string
		method     string
		url        string
		body       string
		expectCode int
	}{
		{
			name:       "register missing email",
			method:     http.MethodPost,
			url:        "/api/v1/auth/register",
			body:       `{"name":"Bob","password":"password123"}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "register invalid email format",
			method:     http.MethodPost,
			url:        "/api/v1/auth/register",
			body:       `{"name":"Bob","email":"not-an-email","password":"password123"}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "register short password (< 6 chars)",
			method:     http.MethodPost,
			url:        "/api/v1/auth/register",
			body:       `{"name":"Bob","email":"bob@example.com","password":"123"}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "register empty name",
			method:     http.MethodPost,
			url:        "/api/v1/auth/register",
			body:       `{"name":"","email":"bob@example.com","password":"password123"}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "register invalid json payload",
			method:     http.MethodPost,
			url:        "/api/v1/auth/register",
			body:       `{"name":"Bob", email: invalid}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "login missing password",
			method:     http.MethodPost,
			url:        "/api/v1/auth/login",
			body:       `{"email":"bob@example.com"}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "login invalid email format",
			method:     http.MethodPost,
			url:        "/api/v1/auth/login",
			body:       `{"email":"invalid","password":"password123"}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "login malformed json",
			method:     http.MethodPost,
			url:        "/api/v1/auth/login",
			body:       `not-json`,
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectCode {
				t.Fatalf("expected HTTP %d, got %d: %s", tc.expectCode, w.Code, w.Body.String())
			}
		})
	}
}
