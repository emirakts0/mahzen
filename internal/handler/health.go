package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthStatus represents the health status of individual components.
type HealthStatus struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Count   int64  `json:"count,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse is the response for the health check endpoint.
type HealthResponse struct {
	Status       string       `json:"status"`
	Database     HealthStatus `json:"database"`
	SearchEngine HealthStatus `json:"search_engine"`
}

// healthHandler handles health check requests.
type healthHandler struct {
	dbPing         func(context.Context) error
	dbCount        func(context.Context) (int64, error)
	searchHealth   func(context.Context, time.Duration) (bool, error)
	searchDocCount func(context.Context) (int64, error)
}

// newHealthHandler creates a new health handler.
func newHealthHandler(
	dbPing func(context.Context) error,
	dbCount func(context.Context) (int64, error),
	searchHealth func(context.Context, time.Duration) (bool, error),
	searchDocCount func(context.Context) (int64, error),
) *healthHandler {
	return &healthHandler{
		dbPing:         dbPing,
		dbCount:        dbCount,
		searchHealth:   searchHealth,
		searchDocCount: searchDocCount,
	}
}

func (h *healthHandler) check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp := HealthResponse{
		Status:       "ok",
		Database:     h.checkDatabase(ctx),
		SearchEngine: h.checkSearchEngine(ctx),
	}

	// If any component is unhealthy, return 503.
	if resp.Database.Status != "healthy" || resp.SearchEngine.Status != "healthy" {
		resp.Status = "degraded"
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *healthHandler) checkDatabase(ctx context.Context) HealthStatus {
	if h.dbPing == nil {
		return HealthStatus{Status: "unknown", Error: "ping function not configured"}
	}

	start := time.Now()
	if err := h.dbPing(ctx); err != nil {
		return HealthStatus{Status: "unhealthy", Error: err.Error()}
	}

	status := HealthStatus{Status: "healthy", Latency: time.Since(start).String()}

	// Get entry count if available.
	if h.dbCount != nil {
		if count, err := h.dbCount(ctx); err == nil {
			status.Count = count
		}
	}

	return status
}

func (h *healthHandler) checkSearchEngine(ctx context.Context) HealthStatus {
	if h.searchHealth == nil {
		return HealthStatus{Status: "unknown", Error: "health function not configured"}
	}

	start := time.Now()
	ok, err := h.searchHealth(ctx, 3*time.Second)
	if err != nil {
		return HealthStatus{Status: "unhealthy", Error: err.Error()}
	}
	if !ok {
		return HealthStatus{Status: "unhealthy", Error: "health check returned false"}
	}

	status := HealthStatus{Status: "healthy", Latency: time.Since(start).String()}

	// Get document count if available.
	if h.searchDocCount != nil {
		if count, err := h.searchDocCount(ctx); err == nil {
			status.Count = count
		}
	}

	return status
}
