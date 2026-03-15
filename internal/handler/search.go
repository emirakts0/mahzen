package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service"
)

// searchHandler implements the search HTTP handlers.
type searchHandler struct {
	svc *service.SearchService
}

// newSearchHandler creates a new searchHandler.
func newSearchHandler(svc *service.SearchService) *searchHandler {
	return &searchHandler{svc: svc}
}

// highlightResponse is the JSON representation of a field-attributed highlight.
type highlightResponse struct {
	Field   string `json:"field"`
	Snippet string `json:"snippet"`
}

// searchResultResponse is the JSON representation of a search result.
type searchResultResponse struct {
	EntryID    string              `json:"entry_id"`
	IsMine     bool                `json:"is_mine"`
	Title      string              `json:"title"`
	Summary    string              `json:"summary,omitempty"`
	Content    string              `json:"content,omitempty"`
	Score      float64             `json:"score,omitempty"`
	Highlights []highlightResponse `json:"highlights,omitempty"`
	Path       string              `json:"path"`
	Visibility string              `json:"visibility"`
	Tags       []string            `json:"tags,omitempty"`
	CreatedAt  string              `json:"created_at"`
	FileType   string              `json:"file_type,omitempty"`
	FileSize   int64               `json:"file_size,omitempty"`
}

func (h *searchHandler) keywordSearch(c *gin.Context) {
	query := c.Query("query")
	userID := userIDFromContext(c)
	filters := parseSearchFilters(c)

	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	results, total, err := h.svc.KeywordSearch(c.Request.Context(), query, userID, filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "keyword search: " + err.Error()})
		return
	}

	items := make([]searchResultResponse, len(results))
	for i, r := range results {
		highlights := make([]highlightResponse, len(r.Highlights))
		for j, h := range r.Highlights {
			highlights[j] = highlightResponse{Field: h.Field, Snippet: h.Snippet}
		}
		items[i] = searchResultResponse{
			EntryID:    r.EntryID,
			IsMine:     r.UserID == userID,
			Title:      r.Title,
			Summary:    r.Summary,
			Content:    r.Content,
			Highlights: highlights,
			Path:       r.Path,
			Visibility: r.Visibility,
			Tags:       r.Tags,
			CreatedAt:  r.CreatedAt,
			FileType:   r.FileType,
			FileSize:   r.FileSize,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results": items,
		"total":   total,
	})
}

func (h *searchHandler) semanticSearch(c *gin.Context) {
	query := c.Query("query")
	userID := userIDFromContext(c)
	filters := parseSearchFilters(c)

	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	results, total, err := h.svc.SemanticSearch(c.Request.Context(), query, userID, filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "semantic search: " + err.Error()})
		return
	}

	items := make([]searchResultResponse, len(results))
	for i, r := range results {
		highlights := make([]highlightResponse, len(r.Highlights))
		for j, h := range r.Highlights {
			highlights[j] = highlightResponse{Field: h.Field, Snippet: h.Snippet}
		}
		items[i] = searchResultResponse{
			EntryID:    r.EntryID,
			IsMine:     r.UserID == userID,
			Title:      r.Title,
			Summary:    r.Summary,
			Content:    r.Content,
			Score:      r.Score,
			Highlights: highlights,
			Path:       r.Path,
			Visibility: r.Visibility,
			Tags:       r.Tags,
			CreatedAt:  r.CreatedAt,
			FileType:   r.FileType,
			FileSize:   r.FileSize,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results": items,
		"total":   total,
	})
}

// parseSearchFilters extracts optional search filters from query parameters.
// Supported params: tags (comma-separated), path, from_date, to_date, only_mine, visibility
func parseSearchFilters(c *gin.Context) *domain.SearchFilters {
	filters := &domain.SearchFilters{}

	// Tags: comma-separated list (e.g., ?tags=go,rust,web)
	if tags := c.Query("tags"); tags != "" {
		for _, tag := range strings.Split(tags, ",") {
			if t := strings.TrimSpace(tag); t != "" {
				filters.Tags = append(filters.Tags, t)
			}
		}
	}

	// Path: exact path match (e.g., ?path=/notes/work)
	if path := c.Query("path"); path != "" {
		filters.Path = strings.TrimSpace(path)
	}

	// FromDate: ISO date format (e.g., ?from_date=2024-01-01)
	if fromDate := c.Query("from_date"); fromDate != "" {
		if t, err := time.Parse(time.DateOnly, fromDate); err == nil {
			filters.FromDate = t
		}
	}

	// ToDate: ISO date format (e.g., ?to_date=2024-12-31)
	if toDate := c.Query("to_date"); toDate != "" {
		if t, err := time.Parse(time.DateOnly, toDate); err == nil {
			filters.ToDate = t
		}
	}

	// OnlyMine: boolean (e.g., ?only_mine=true)
	if onlyMine := c.Query("only_mine"); onlyMine != "" {
		filters.OnlyMine = strings.ToLower(onlyMine) == "true" || onlyMine == "1"
	}

	// Visibility: "public" or "private" (e.g., ?visibility=private)
	if visibility := c.Query("visibility"); visibility != "" {
		v := strings.ToLower(strings.TrimSpace(visibility))
		if v == "public" || v == "private" {
			filters.Visibility = v
		}
	}

	// Return nil if no filters were set
	if len(filters.Tags) == 0 &&
		filters.Path == "" &&
		filters.FromDate.IsZero() &&
		filters.ToDate.IsZero() &&
		!filters.OnlyMine &&
		filters.Visibility == "" {
		return nil
	}

	return filters
}
