package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// respondData writes a JSON response.
func respondData(c *gin.Context, code int, data any) {
	c.JSON(code, data)
}

// respondError writes a JSON error response with an "error" field.
func respondError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}

// bindJSON parses the request body into dst, responding with 400 on failure.
func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

// requireUser returns the authenticated user ID, responding with 401 when absent.
func requireUser(c *gin.Context) (string, bool) {
	userID := userIDFromContext(c)
	if userID == "" {
		respondError(c, http.StatusUnauthorized, "user not authenticated")
		return "", false
	}
	return userID, true
}

// parsePagination reads limit/offset query parameters with a default limit.
func parsePagination(c *gin.Context, defaultLimit int) (limit, offset int) {
	limit = defaultLimit
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = v
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v >= 0 {
		offset = v
	}
	return limit, offset
}

// parseDateFilter parses optional from_date/to_date ISO date parameters.
// Unparseable values are silently ignored.
func parseDateFilter(fromDate, toDate string) (from, to time.Time) {
	if t, err := time.Parse(time.DateOnly, fromDate); err == nil {
		from = t
	}
	if t, err := time.Parse(time.DateOnly, toDate); err == nil {
		to = t
	}
	return from, to
}
