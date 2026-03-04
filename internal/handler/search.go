package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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

// searchResultResponse is the JSON representation of a search result.
type searchResultResponse struct {
	EntryID    string   `json:"entry_id"`
	Title      string   `json:"title"`
	Snippet    string   `json:"snippet"`
	Score      float64  `json:"score,omitempty"`
	Highlights []string `json:"highlights,omitempty"`
	Path       string   `json:"path"`
	Visibility string   `json:"visibility"`
	Tags       []string `json:"tags,omitempty"`
	CreatedAt  string   `json:"created_at"`
}

func (h *searchHandler) keywordSearch(c *gin.Context) {
	query := c.Query("query")
	userID := userIDFromContext(c)

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

	results, total, err := h.svc.KeywordSearch(c.Request.Context(), query, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "keyword search: " + err.Error()})
		return
	}

	items := make([]searchResultResponse, len(results))
	for i, r := range results {
		items[i] = searchResultResponse{
			EntryID:    r.EntryID,
			Title:      r.Title,
			Snippet:    r.Snippet,
			Highlights: r.Highlights,
			Path:       r.Path,
			Visibility: r.Visibility,
			Tags:       r.Tags,
			CreatedAt:  r.CreatedAt,
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

	results, total, err := h.svc.SemanticSearch(c.Request.Context(), query, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "semantic search: " + err.Error()})
		return
	}

	items := make([]searchResultResponse, len(results))
	for i, r := range results {
		items[i] = searchResultResponse{
			EntryID:    r.EntryID,
			Title:      r.Title,
			Snippet:    r.Snippet,
			Score:      r.Score,
			Highlights: r.Highlights,
			Path:       r.Path,
			Visibility: r.Visibility,
			Tags:       r.Tags,
			CreatedAt:  r.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results": items,
		"total":   total,
	})
}
