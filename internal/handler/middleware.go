package handler

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/emirakts0/mahzen/internal/domain"
)

const userIDKey = "user_id"

// userIDFromContext extracts the user ID set by AuthMiddleware.
func userIDFromContext(c *gin.Context) string {
	return c.GetString(userIDKey)
}

// publicPaths lists routes that work without authentication.
var publicPaths = map[string]bool{
	"GET /health":                 true,
	"POST /v1/auth/register":      true,
	"POST /v1/auth/login":         true,
	"POST /v1/auth/refresh":       true,
	"POST /v1/auth/logout":        true,
	"GET /v1/entries":             true,
	"GET /v1/search/keyword":      true,
	"GET /v1/search/semantic":     true,
	"GET /v1/downloads/:platform": true,
}

func isPublicRoute(method, path string) bool {
	// Exact match first.
	key := method + " " + path
	if publicPaths[key] {
		return true
	}
	// GET /v1/entries/:id is also public.
	if method == "GET" && strings.HasPrefix(path, "/v1/entries/") {
		return true
	}
	return false
}

// AuthMiddleware returns a Gin middleware that validates JWT access tokens
// and opaque "mah_" tokens, setting the user ID in the request context.
// Public routes work without authentication; a valid token still scopes
// results to the caller when provided.
func AuthMiddleware(logger *slog.Logger, tokenGen domain.TokenGenerator, tokenStore domain.AccessTokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Dev bypass: a raw user ID header skips token validation.
		if uid := c.GetHeader("X-User-Id"); uid != "" {
			c.Set(userIDKey, uid)
			c.Next()
			return
		}

		isPublic := isPublicRoute(c.Request.Method, c.FullPath())
		unauthorized := func(msg string) {
			if isPublic {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			unauthorized("authentication required")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		var userID string
		switch {
		// Opaque tokens start with "mah_" — skip JWT parsing.
		case strings.HasPrefix(token, "mah_"):
			userID, _ = tokenStore.Lookup(token)
		default:
			if id, err := tokenGen.ValidateAccessToken(token); err != nil {
				logger.Debug("jwt validation failed",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"error", err,
				)
			} else {
				userID = id
			}
		}

		if userID == "" {
			unauthorized("invalid or expired token")
			return
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}

func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		logArgs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"duration", duration,
			"client_ip", c.ClientIP(),
		}

		if len(c.Errors) > 0 {
			logArgs = append(logArgs, "errors", c.Errors.String())
		}

		logger.Info("http request", logArgs...)
	}
}

func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"panic", r,
					"stack", string(debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://localhost:5173": true,
		"http://127.0.0.1:3000": true,
		"http://127.0.0.1:5173": true,
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-Id")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
