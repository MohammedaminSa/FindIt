package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockPinger implements Pinger interface for unit tests
type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	return m.err
}

func TestHealthCheck_Healthy(t *testing.T) {
	mockDB := &mockPinger{err: nil}
	h := NewHealthHandler(mockDB)

	router := gin.New()
	router.GET("/api/v1/health", h.Check)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var res HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Status != "UP" {
		t.Errorf("expected status 'UP', got '%s'", res.Status)
	}

	if res.Services["database"] != "UP" {
		t.Errorf("expected database service 'UP', got '%s'", res.Services["database"])
	}

	if res.Uptime == "" {
		t.Error("expected non-empty uptime")
	}

	if res.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestHealthCheck_Degraded(t *testing.T) {
	mockDB := &mockPinger{err: errors.New("connection refused")}
	h := NewHealthHandler(mockDB)

	router := gin.New()
	router.GET("/api/v1/health", h.Check)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var res HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Status != "DEGRADED" {
		t.Errorf("expected status 'DEGRADED', got '%s'", res.Status)
	}

	if res.Services["database"] != "DOWN" {
		t.Errorf("expected database service 'DOWN', got '%s'", res.Services["database"])
	}
}

func TestHealthCheck_NoDB(t *testing.T) {
	h := NewHealthHandler(nil)

	router := gin.New()
	router.GET("/api/v1/health", h.Check)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var res HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Status != "UP" {
		t.Errorf("expected status 'UP', got '%s'", res.Status)
	}

	if res.Services["database"] != "NOT_CONFIGURED" {
		t.Errorf("expected database service 'NOT_CONFIGURED', got '%s'", res.Services["database"])
	}
}
