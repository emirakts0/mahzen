package handler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/emirakts0/mahzen/internal/domain"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ---------------------------------------------------------------------------
// domainEntryToResponse
// ---------------------------------------------------------------------------

func TestDomainEntryToResponse(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	entry := &domain.Entry{
		ID:         "entry-1",
		UserID:     "user-1",
		Title:      "Test Entry",
		Content:    "some content",
		Summary:    "a summary",
		Path:       "/notes/work",
		Visibility: domain.VisibilityPublic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	resp := domainEntryToResponse(entry, []string{"tag-1", "tag-2"})
	assert.Equal(t, "entry-1", resp.ID)
	assert.Equal(t, "user-1", resp.UserID)
	assert.Equal(t, "Test Entry", resp.Title)
	assert.Equal(t, "some content", resp.Content)
	assert.Equal(t, "a summary", resp.Summary)
	assert.Equal(t, "/notes/work", resp.Path)
	assert.Equal(t, "public", resp.Visibility)
	assert.Equal(t, []string{"tag-1", "tag-2"}, resp.Tags)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
}

func TestDomainEntryToResponse_NilTags(t *testing.T) {
	entry := &domain.Entry{ID: "e1", Visibility: domain.VisibilityPrivate}
	resp := domainEntryToResponse(entry, nil)
	assert.Nil(t, resp.Tags)
	assert.Equal(t, "private", resp.Visibility)
}

// ---------------------------------------------------------------------------
// domainTagToResponse
// ---------------------------------------------------------------------------

func TestDomainTagToResponse(t *testing.T) {
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	tag := &domain.Tag{
		ID:        "tag-1",
		Name:      "Golang",
		Slug:      "golang",
		CreatedAt: now,
	}

	resp := domainTagToResponse(tag)
	assert.Equal(t, "tag-1", resp.ID)
	assert.Equal(t, "Golang", resp.Name)
	assert.Equal(t, "golang", resp.Slug)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
}

// ---------------------------------------------------------------------------
// Visibility conversion
// ---------------------------------------------------------------------------

func TestParseVisibility(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected domain.Visibility
	}{
		{"public", "public", domain.VisibilityPublic},
		{"private", "private", domain.VisibilityPrivate},
		{"empty defaults to private", "", domain.VisibilityPrivate},
		{"unknown defaults to private", "unknown", domain.VisibilityPrivate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, domain.ParseVisibility(tt.input))
		})
	}
}

func TestVisibilityString(t *testing.T) {
	tests := []struct {
		name     string
		input    domain.Visibility
		expected string
	}{
		{"public", domain.VisibilityPublic, "public"},
		{"private", domain.VisibilityPrivate, "private"},
		{"unknown", domain.Visibility(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.String())
		})
	}
}

// ---------------------------------------------------------------------------
// userIDFromContext
// ---------------------------------------------------------------------------

func TestUserIDFromContext_Present(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(userIDKey, "user-42")

	assert.Equal(t, "user-42", userIDFromContext(c))
}

func TestUserIDFromContext_Missing(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	assert.Empty(t, userIDFromContext(c))
}

// ---------------------------------------------------------------------------
// LoggingMiddleware
// ---------------------------------------------------------------------------

func TestLoggingMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.Use(LoggingMiddleware(logger))
	engine.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

// ---------------------------------------------------------------------------
// RecoveryMiddleware
// ---------------------------------------------------------------------------

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.Use(RecoveryMiddleware(logger))
	engine.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestRecoveryMiddleware_Panic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.Use(RecoveryMiddleware(logger))
	engine.GET("/test", func(c *gin.Context) {
		panic("unexpected error")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
