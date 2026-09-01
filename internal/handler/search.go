package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service"
)

// searchHandler implements the search HTTP handlers.
type searchHandler struct {
	svc *service.SearchService
}

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

func domainSearchResultsToResponses(results []*domain.SearchResult, userID string) []searchResultResponse {
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
	return items
}

// runSearch executes a search and renders the shared response shape.
type searchFunc func(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error)

func (h *searchHandler) runSearch(c *gin.Context, kind string, fn searchFunc) {
	userID := userIDFromContext(c)
	limit, offset := parsePagination(c, 20)

	results, total, err := fn(c.Request.Context(), c.Query("query"), userID, parseSearchFilters(c), limit, offset)
	if err != nil {
		respondError(c, http.StatusInternalServerError, kind+" search: "+err.Error())
		return
	}

	respondData(c, http.StatusOK, gin.H{
		"results": domainSearchResultsToResponses(results, userID),
		"total":   total,
	})
}

func (h *searchHandler) keywordSearch(c *gin.Context) {
	h.runSearch(c, "keyword", h.svc.KeywordSearch)
}

func (h *searchHandler) semanticSearch(c *gin.Context) {
	h.runSearch(c, "semantic", h.svc.SemanticSearch)
}

// parseSearchFilters extracts optional search filters from query parameters.
// Supported params: tags, path, from_date, to_date, only_mine, visibility.
// Returns nil when no filter is set.
func parseSearchFilters(c *gin.Context) *domain.SearchFilters {
	filters := &domain.SearchFilters{}

	if tags := c.Query("tags"); tags != "" {
		for tag := range strings.SplitSeq(tags, ",") {
			if t := strings.TrimSpace(tag); t != "" {
				filters.Tags = append(filters.Tags, t)
			}
		}
	}
	filters.Path = strings.TrimSpace(c.Query("path"))
	filters.FromDate, filters.ToDate = parseDateFilter(c.Query("from_date"), c.Query("to_date"))
	filters.OnlyMine = c.Query("only_mine") == "true" || c.Query("only_mine") == "1"
	if v := strings.ToLower(strings.TrimSpace(c.Query("visibility"))); v == "public" || v == "private" {
		filters.Visibility = v
	}

	if filters.Tags == nil && filters.Path == "" && filters.FromDate.IsZero() &&
		filters.ToDate.IsZero() && !filters.OnlyMine && filters.Visibility == "" {
		return nil
	}
	return filters
}
