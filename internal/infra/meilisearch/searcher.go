package meilisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	meili "github.com/meilisearch/meilisearch-go"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/infra/searchutil"
)

// Searcher implements domain.Searcher using Meilisearch.
type Searcher struct {
	client meili.IndexManager
}

// NewSearcher creates a new Meilisearch-backed searcher.
func NewSearcher(client meili.ServiceManager) *Searcher {
	return &Searcher{client: client.Index(IndexName)}
}

func (s *Searcher) KeywordSearch(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
	filter := buildFilter(userID, filters)

	slog.Info("meilisearch keyword search",
		"query", query,
		"user_id", userID,
		"filter", filter,
		"limit", limit,
		"offset", offset,
	)

	start := time.Now()
	resp, err := s.client.SearchWithContext(ctx, query, &meili.SearchRequest{
		Filter:                filter,
		Limit:                 int64(limit),
		Offset:                int64(offset),
		AttributesToHighlight: []string{"title", "content", "summary"},
		AttributesToCrop:      []string{"content"},
		CropLength:            30,
		CropMarker:            "…",
		HighlightPreTag:       "<mark>",
		HighlightPostTag:      "</mark>",
		MatchingStrategy:      meili.Last,
		ShowRankingScore:      true,
	})
	duration := time.Since(start)

	if err != nil {
		slog.Error("meilisearch keyword search failed",
			"query", query,
			"duration", duration,
			"error", err,
		)
		return nil, 0, fmt.Errorf("keyword search: %w", err)
	}

	results, total := mapSearchResults(resp)

	slog.Info("meilisearch keyword search completed",
		"query", query,
		"duration", duration,
		"total_found", total,
		"returned", len(results),
	)

	return results, total, nil
}

func (s *Searcher) SemanticSearch(ctx context.Context, embedding []float32, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
	filter := buildFilter(userID, filters)

	slog.Info("meilisearch semantic search",
		"user_id", userID,
		"filter", filter,
		"embedding_dims", len(embedding),
		"limit", limit,
		"offset", offset,
	)

	start := time.Now()
	resp, err := s.client.SearchWithContext(ctx, "", &meili.SearchRequest{
		Filter:           filter,
		Limit:            int64(limit),
		Offset:           int64(offset),
		Vector:           embedding,
		Hybrid:           &meili.SearchRequestHybrid{Embedder: "openai", SemanticRatio: 1.0},
		ShowRankingScore: true,
	})
	duration := time.Since(start)

	if err != nil {
		slog.Error("meilisearch semantic search failed",
			"duration", duration,
			"error", err,
		)
		return nil, 0, fmt.Errorf("semantic search: %w", err)
	}

	results, total := mapSearchResults(resp)

	slog.Info("meilisearch semantic search completed",
		"duration", duration,
		"total_found", total,
		"returned", len(results),
	)

	return results, total, nil
}

// buildFilter creates a Meilisearch filter expression that enforces visibility
// rules and applies optional filters for tags, path, date range, etc.
func buildFilter(userID string, filters *domain.SearchFilters) string {
	conditions := make([]string, 0, 6)

	// Visibility filter with security
	if filters != nil && filters.Visibility == "public" {
		conditions = append(conditions, `visibility = "public"`)
	} else if filters != nil && filters.Visibility == "private" {
		if userID == "" {
			conditions = append(conditions, `visibility = "__impossible__"`)
		} else {
			conditions = append(conditions, fmt.Sprintf(`(visibility = "private" AND user_id = "%s")`, userID))
		}
	} else if userID == "" {
		conditions = append(conditions, `visibility = "public"`)
	} else {
		conditions = append(conditions, fmt.Sprintf(`(visibility = "public" OR user_id = "%s")`, userID))
	}

	if filters == nil {
		return strings.Join(conditions, " AND ")
	}

	if filters.OnlyMine && userID != "" {
		conditions = append(conditions, fmt.Sprintf(`user_id = "%s"`, userID))
	}

	// Tags filter (OR logic)
	if len(filters.Tags) > 0 {
		escapedTags := make([]string, len(filters.Tags))
		for i, tag := range filters.Tags {
			escapedTags[i] = escapeFilterValue(tag)
		}
		conditions = append(conditions, fmt.Sprintf("tags IN [%s]", strings.Join(escapedTags, ", ")))
	}

	// Path filter (exact match)
	if filters.Path != "" {
		conditions = append(conditions, fmt.Sprintf(`path = "%s"`, escapeQuotes(filters.Path)))
	}

	// Date range filters
	if !filters.FromDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at >= %d", filters.FromDate.Unix()))
	}
	if !filters.ToDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at <= %d", filters.ToDate.Unix()))
	}

	return strings.Join(conditions, " AND ")
}

func escapeFilterValue(val string) string {
	return `"` + escapeQuotes(val) + `"`
}

func escapeQuotes(val string) string {
	return strings.ReplaceAll(val, `"`, `\"`)
}

// mapSearchResults converts a Meilisearch search response to domain search results.
func mapSearchResults(resp *meili.SearchResponse) ([]*domain.SearchResult, int) {
	if resp.Hits == nil || len(resp.Hits) == 0 {
		return nil, 0
	}

	total := int(resp.EstimatedTotalHits)

	results := make([]*domain.SearchResult, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		sr := &domain.SearchResult{}

		// Extract regular fields from the hit.
		sr.EntryID = stringFromHit(hit, "id")
		sr.UserID = stringFromHit(hit, "user_id")
		sr.Title = stringFromHit(hit, "title")
		sr.Summary = stringFromHit(hit, "summary")
		sr.Path = stringFromHit(hit, "path")
		sr.Visibility = stringFromHit(hit, "visibility")
		sr.FileType = stringFromHit(hit, "file_type")
		sr.Tags = stringsFromHit(hit, "tags")
		sr.FileSize = int64FromHit(hit, "file_size")

		// Timestamp.
		if ts := int64FromHit(hit, "created_at"); ts > 0 {
			sr.CreatedAt = time.Unix(ts, 0).UTC().Format(time.RFC3339)
		}

		// Ranking score.
		if score := float64FromHit(hit, "_rankingScore"); score > 0 {
			sr.Score = score
		}

		// Highlights and cropped content from _formatted.
		if formatted := objectFromHit(hit, "_formatted"); formatted != nil {
			// Use cropped content from Meilisearch (centered around match).
			if cropped := stringFromMap(formatted, "content"); cropped != "" {
				sr.Content = cropped
			}

			for _, field := range []string{"title", "content", "summary"} {
				if snippet := stringFromMap(formatted, field); snippet != "" && strings.Contains(snippet, "<mark>") {
					sr.Highlights = append(sr.Highlights, domain.Highlight{
						Field:   field,
						Snippet: snippet,
					})
				}
			}
		}

		// Fallback: use raw content if _formatted didn't provide cropped content.
		if sr.Content == "" {
			if raw := stringFromHit(hit, "content"); raw != "" {
				if len(raw) <= searchutil.ContentExcerptLen {
					sr.Content = raw
				} else {
					sr.Content = raw[:searchutil.ContentExcerptLen]
				}
			}
		}

		results = append(results, sr)
	}

	return results, total
}

func stringFromHit(hit meili.Hit, key string) string {
	raw, ok := hit[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return s
}

func stringsFromHit(hit meili.Hit, key string) []string {
	raw, ok := hit[key]
	if !ok {
		return nil
	}
	var ss []string
	if err := json.Unmarshal(raw, &ss); err != nil {
		return nil
	}
	return ss
}

func int64FromHit(hit meili.Hit, key string) int64 {
	raw, ok := hit[key]
	if !ok {
		return 0
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err != nil {
		return 0
	}
	i, err := n.Int64()
	if err != nil {
		return 0
	}
	return i
}

func float64FromHit(hit meili.Hit, key string) float64 {
	raw, ok := hit[key]
	if !ok {
		return 0
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return 0
	}
	return f
}

func objectFromHit(hit meili.Hit, key string) map[string]json.RawMessage {
	raw, ok := hit[key]
	if !ok {
		return nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}

func stringFromMap(m map[string]json.RawMessage, key string) string {
	raw, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return s
}
