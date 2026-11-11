package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/handler"
	"github.com/MohammedaminSa/FindIt/backend/internal/middleware"
	"github.com/MohammedaminSa/FindIt/backend/pkg/token"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSetupRouter_HealthEndpoint(t *testing.T) {
	router := setupRouter(nil, nil)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	if err != nil {
		t.Fatalf("failed to create test request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var res handler.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if res.Status != "UP" {
		t.Errorf("expected status 'UP', got '%s'", res.Status)
	}

	if res.Services["database"] != "NOT_CONFIGURED" {
		t.Errorf("expected database service 'NOT_CONFIGURED', got '%s'", res.Services["database"])
	}
}

func TestSetupRouter_AuthEndpointsMounted(t *testing.T) {
	jwtMgr, err := token.NewJWTManager("test-jwt-secret-key-at-least-32-chars-long", time.Hour)
	if err != nil {
		t.Fatalf("failed to initialize jwt manager: %v", err)
	}
	authMw := middleware.Auth(jwtMgr)
	authHandler := handler.NewAuthHandler(nil)
	router := setupRouter(nil, authHandler, authMw)

	// Test that /api/v1/auth/register responds (not 404)
	reqReg, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	wReg := httptest.NewRecorder()
	router.ServeHTTP(wReg, reqReg)

	if wReg.Code == http.StatusNotFound {
		t.Errorf("expected /api/v1/auth/register to be mounted, got 404")
	}

	// Test that /api/v1/auth/login responds (not 404)
	reqLogin, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code == http.StatusNotFound {
		t.Errorf("expected /api/v1/auth/login to be mounted, got 404")
	}

	// Test that /api/v1/auth/me is mounted and protected by auth middleware
	reqMe, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	wMe := httptest.NewRecorder()
	router.ServeHTTP(wMe, reqMe)

	if wMe.Code != http.StatusUnauthorized {
		t.Errorf("expected /api/v1/auth/me to return 401 without auth header, got %d", wMe.Code)
	}
}
