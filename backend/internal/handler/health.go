package handler

import (
	"context"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
)

// Pinger defines an interface for components capable of verifying database connectivity.
// This interface allows seamless mocking in tests without needing a running database.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler manages system health probes and connectivity checks.
type HealthHandler struct {
	db        Pinger
	startTime time.Time
}

// NewHealthHandler constructs a HealthHandler with an optional database pinger and records server startup time.
func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
	}
}

// isNilPinger safely checks if the pinger interface or its underlying concrete pointer is nil.
func isNilPinger(p Pinger) bool {
	if p == nil {
		return true
	}
	v := reflect.ValueOf(p)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

// HealthResponse represents the structured JSON payload returned by the health endpoint.
type HealthResponse struct {
	Status    string            `json:"status"`
	Uptime    string            `json:"uptime"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// Check responds to GET /api/v1/health with current uptime and service dependency status.
func (h *HealthHandler) Check(c *gin.Context) {
	services := make(map[string]string)
	overallStatus := "UP"
	httpStatusCode := http.StatusOK

	if !isNilPinger(h.db) {
		// Use a dedicated 2-second timeout context for the health ping
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := h.db.Ping(ctx); err != nil {
			services["database"] = "DOWN"
			overallStatus = "DEGRADED"
			httpStatusCode = http.StatusServiceUnavailable
		} else {
			services["database"] = "UP"
		}
	} else {
		services["database"] = "NOT_CONFIGURED"
	}

	c.JSON(httpStatusCode, HealthResponse{
		Status:    overallStatus,
		Uptime:    time.Since(h.startTime).Truncate(time.Second).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	})
}
