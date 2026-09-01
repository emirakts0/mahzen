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
	visibility := visibilityCondition(userID, filters)
	if filters == nil {
		return visibility
	}

	conditions := []string{visibility}

	if filters.OnlyMine && userID != "" {
		conditions = append(conditions, fmt.Sprintf(`user_id = "%s"`, userID))
	}
	if len(filters.Tags) > 0 {
		escaped := make([]string, len(filters.Tags))
		for i, tag := range filters.Tags {
			escaped[i] = escapeFilterValue(tag)
		}
		conditions = append(conditions, fmt.Sprintf("tags IN [%s]", strings.Join(escaped, ", ")))
	}
	if filters.Path != "" {
		conditions = append(conditions, fmt.Sprintf(`path = %s`, escapeFilterValue(filters.Path)))
	}
	if !filters.FromDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at >= %d", filters.FromDate.Unix()))
	}
	if !filters.ToDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at <= %d", filters.ToDate.Unix()))
	}

	return strings.Join(conditions, " AND ")
}

// visibilityCondition enforces entry visibility: anonymous users only ever
// see public entries; a "private" filter without a user yields an impossible
// condition so it matches nothing.
func visibilityCondition(userID string, filters *domain.SearchFilters) string {
	switch {
	case filters != nil && filters.Visibility == "public":
		return `visibility = "public"`
	case filters != nil && filters.Visibility == "private":
		if userID == "" {
			return `visibility = "__impossible__"`
		}
		return fmt.Sprintf(`(visibility = "private" AND user_id = "%s")`, userID)
	case userID == "":
		return `visibility = "public"`
	default:
		return fmt.Sprintf(`(visibility = "public" OR user_id = "%s")`, userID)
	}
}

func escapeFilterValue(val string) string {
	return `"` + strings.ReplaceAll(val, `"`, `\"`) + `"`
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

		sr.EntryID = stringFromHit(hit, "id")
		sr.UserID = stringFromHit(hit, "user_id")
		sr.Title = stringFromHit(hit, "title")
		sr.Summary = stringFromHit(hit, "summary")
		sr.Path = stringFromHit(hit, "path")
		sr.Visibility = stringFromHit(hit, "visibility")
		sr.FileType = stringFromHit(hit, "file_type")
		sr.Tags = stringsFromHit(hit, "tags")
		sr.FileSize = int64FromHit(hit, "file_size")

		if ts := int64FromHit(hit, "created_at"); ts > 0 {
			sr.CreatedAt = time.Unix(ts, 0).UTC().Format(time.RFC3339)
		}

		if score := float64FromHit(hit, "_rankingScore"); score > 0 {
			sr.Score = score
		}

		if formatted := objectFromHit(hit, "_formatted"); formatted != nil {
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

		if sr.Content == "" {
			if raw := stringFromHit(hit, "content"); raw != "" {
				sr.Content = raw[:min(len(raw), searchutil.ContentExcerptLen)]
			}
		}

		results = append(results, sr)
	}

	return results, total
}

// hitValue unmarshals the raw JSON value stored under key in a hit,
// returning the zero value of T on absence or type mismatch.
func hitValue[T any](hit meili.Hit, key string) T {
	var out T
	raw, ok := hit[key]
	if !ok {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}

func stringFromHit(hit meili.Hit, key string) string {
	return hitValue[string](hit, key)
}

func stringsFromHit(hit meili.Hit, key string) []string {
	return hitValue[[]string](hit, key)
}

func int64FromHit(hit meili.Hit, key string) int64 {
	n, _ := hitValue[json.Number](hit, key).Int64()
	return n
}

func float64FromHit(hit meili.Hit, key string) float64 {
	return hitValue[float64](hit, key)
}

func objectFromHit(hit meili.Hit, key string) map[string]json.RawMessage {
	return hitValue[map[string]json.RawMessage](hit, key)
}

func stringFromMap(m map[string]json.RawMessage, key string) string {
	var out string
	raw, ok := m[key]
	if !ok {
		return ""
	}
	_ = json.Unmarshal(raw, &out)
	return out
}
