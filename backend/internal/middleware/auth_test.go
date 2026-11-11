package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAuthMiddlewareRouter(tokenManager token.TokenManager) *gin.Engine {
	router := gin.New()
	router.Use(Auth(tokenManager))
	router.GET("/protected", func(c *gin.Context) {
		userID, userExists := GetUserID(c)
		email, emailExists := GetUserEmail(c)
		claims, claimsExists := GetClaims(c)

		c.JSON(http.StatusOK, gin.H{
			"user_id":       userID.String(),
			"email":         email,
			"user_exists":   userExists,
			"email_exists":  emailExists,
			"claims_exists": claimsExists,
			"claims_sub":    claims.Subject,
		})
	})
	return router
}

func TestAuth_Success(t *testing.T) {
	secret := "test-super-secret-key-that-is-at-least-32-bytes-long"
	jwtMgr, err := token.NewJWTManager(secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	userID := uuid.New()
	email := "user@example.com"
	tok, err := jwtMgr.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	router := setupAuthMiddlewareRouter(jwtMgr)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if resp["user_id"] != userID.String() {
		t.Errorf("expected user_id %s, got %v", userID.String(), resp["user_id"])
	}
	if resp["email"] != email {
		t.Errorf("expected email %s, got %v", email, resp["email"])
	}
	if resp["user_exists"] != true || resp["email_exists"] != true || resp["claims_exists"] != true {
		t.Errorf("expected all context flags to be true: %v", resp)
	}
}

func TestAuth_CaseInsensitiveBearer(t *testing.T) {
	secret := "test-super-secret-key-that-is-at-least-32-bytes-long"
	jwtMgr, _ := token.NewJWTManager(secret, 1*time.Hour)

	userID := uuid.New()
	tok, _ := jwtMgr.GenerateToken(userID, "bearer_test@example.com")

	router := setupAuthMiddlewareRouter(jwtMgr)

	testCases := []string{
		"bearer " + tok,
		"BEARER " + tok,
		"Bearer " + tok,
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", tc)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected HTTP 200 for auth header '%s', got %d", tc, w.Code)
		}
	}
}

func TestAuth_MissingHeader(t *testing.T) {
	secret := "test-super-secret-key-that-is-at-least-32-bytes-long"
	jwtMgr, _ := token.NewJWTManager(secret, 1*time.Hour)
	router := setupAuthMiddlewareRouter(jwtMgr)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected HTTP 401, got %d", w.Code)
	}

	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "authorization header is required" {
		t.Errorf("expected 'authorization header is required', got '%s'", resp["error"])
	}
}

func TestAuth_InvalidHeaderFormat(t *testing.T) {
	secret := "test-super-secret-key-that-is-at-least-32-bytes-long"
	jwtMgr, _ := token.NewJWTManager(secret, 1*time.Hour)
	router := setupAuthMiddlewareRouter(jwtMgr)

	testCases := []struct {
		name   string
		header string
	}{
		{"Single word", "Bearer"},
		{"Three words", "Bearer token extra"},
		{"Whitespace only", "    "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tc.header)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected HTTP 401, got %d", w.Code)
			}

			var resp map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["error"] != "invalid authorization header format" && resp["error"] != "authorization header is required" {
				t.Errorf("unexpected error message: '%s'", resp["error"])
			}
		})
	}
}

func TestAuth_UnsupportedAuthType(t *testing.T) {
	secret := "test-super-secret-key-that-is-at-least-32-bytes-long"
	jwtMgr, _ := token.NewJWTManager(secret, 1*time.Hour)
	router := setupAuthMiddlewareRouter(jwtMgr)

	testCases := []string{
		"Basic dXNlcjpwYXNz",
		"Digest token123",
		"Token abcxyz",
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", tc)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected HTTP 401 for header '%s', got %d", tc, w.Code)
		}

		var resp map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["error"] != "unsupported authorization type" {
			t.Errorf("expected 'unsupported authorization type', got '%s'", resp["error"])
		}
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	secret := "test-super-secret-key-that-is-at-least-32-bytes-long"
	// Generate an already expired token with negative duration
	jwtMgr, _ := token.NewJWTManager(secret, -1*time.Hour)

	userID := uuid.New()
	tok, _ := jwtMgr.GenerateToken(userID, "expired@example.com")

	// Validator with standard duration
	validatorMgr, _ := token.NewJWTManager(secret, 1*time.Hour)
	router := setupAuthMiddlewareRouter(validatorMgr)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected HTTP 401, got %d", w.Code)
	}

	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "token has expired" {
		t.Errorf("expected 'token has expired', got '%s'", resp["error"])
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	secret := "test-super-secret-key-that-is-at-least-32-bytes-long"
	jwtMgr, _ := token.NewJWTManager(secret, 1*time.Hour)
	router := setupAuthMiddlewareRouter(jwtMgr)

	testCases := []struct {
		name  string
		token string
	}{
		{"Malformed string", "not.a.valid.jwt"},
		{"Random characters", "abcdef123456"},
		{"Different secret", func() string {
			otherMgr, _ := token.NewJWTManager("other-secret-key-other-secret-key-1234", 1*time.Hour)
			tok, _ := otherMgr.GenerateToken(uuid.New(), "spoof@example.com")
			return tok
		}()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected HTTP 401, got %d", w.Code)
			}

			var resp map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["error"] != "invalid or malformed token" {
				t.Errorf("expected 'invalid or malformed token', got '%s'", resp["error"])
			}
		})
	}
}

func TestContextHelpers_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Context empty
	if _, ok := GetUserID(c); ok {
		t.Errorf("expected GetUserID to return false on empty context")
	}
	if _, ok := GetUserEmail(c); ok {
		t.Errorf("expected GetUserEmail to return false on empty context")
	}
	if _, ok := GetClaims(c); ok {
		t.Errorf("expected GetClaims to return false on empty context")
	}

	// Context with wrong types
	c.Set(ContextUserIDKey, "not-a-uuid")
	c.Set(ContextUserEmailKey, 12345)
	c.Set(ContextClaimsKey, "not-claims")

	if _, ok := GetUserID(c); ok {
		t.Errorf("expected GetUserID to return false for non-UUID value")
	}
	if _, ok := GetUserEmail(c); ok {
		t.Errorf("expected GetUserEmail to return false for non-string value")
	}
	if _, ok := GetClaims(c); ok {
		t.Errorf("expected GetClaims to return false for non-Claims value")
	}
}
