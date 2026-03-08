package typesense

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/typesense/typesense-go/v3/typesense"
	"github.com/typesense/typesense-go/v3/typesense/api"
	"github.com/typesense/typesense-go/v3/typesense/api/pointer"

	"github.com/emirakts0/mahzen/internal/domain"
)

// Searcher implements domain.Searcher using Typesense.
type Searcher struct {
	client *typesense.Client
}

// NewSearcher creates a new Typesense-backed searcher.
func NewSearcher(client *typesense.Client) *Searcher {
	return &Searcher{client: client}
}

func (s *Searcher) KeywordSearch(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
	filterBy := buildFilterBy(userID, filters)

	slog.Info("typesense keyword search",
		"query", query,
		"user_id", userID,
		"filter_by", filterBy,
		"limit", limit,
		"offset", offset,
	)

	params := &api.SearchCollectionParams{
		Q:               pointer.String(query),
		QueryBy:         pointer.String("title,content,summary,tags"),
		FilterBy:        pointer.String(filterBy),
		PerPage:         pointer.Int(limit),
		Page:            pointer.Int((offset / max(limit, 1)) + 1),
		HighlightFields: pointer.String("title,content,summary"),
	}

	start := time.Now()
	result, err := s.client.Collection(CollectionName).Documents().Search(ctx, params)
	duration := time.Since(start)

	if err != nil {
		slog.Error("typesense keyword search failed",
			"query", query,
			"duration", duration,
			"error", err,
		)
		return nil, 0, fmt.Errorf("keyword search: %w", err)
	}

	results, total, mapErr := mapSearchResults(result)

	slog.Info("typesense keyword search completed",
		"query", query,
		"duration", duration,
		"total_found", total,
		"returned", len(results),
	)

	return results, total, mapErr
}

func (s *Searcher) SemanticSearch(ctx context.Context, embedding []float32, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
	filterBy := buildFilterBy(userID, filters)
	vectorQuery := fmt.Sprintf("embedding:([%s], k:%d)", floatsToString(embedding), max(limit+offset, limit))

	slog.Info("typesense semantic search",
		"user_id", userID,
		"filter_by", filterBy,
		"embedding_dims", len(embedding),
		"limit", limit,
		"offset", offset,
	)

	collection := CollectionName
	q := "*"
	perPage := limit
	pageNum := (offset / max(limit, 1)) + 1
	excludeFields := "embedding"

	start := time.Now()
	result, err := s.client.MultiSearch.Perform(ctx, &api.MultiSearchParams{}, api.MultiSearchSearchesParameter{
		Searches: []api.MultiSearchCollectionParameters{
			{
				Collection:    &collection,
				Q:             &q,
				VectorQuery:   &vectorQuery,
				FilterBy:      &filterBy,
				PerPage:       &perPage,
				Page:          &pageNum,
				ExcludeFields: &excludeFields,
			},
		},
	})
	duration := time.Since(start)

	if err != nil {
		slog.Error("typesense semantic search failed",
			"duration", duration,
			"error", err,
		)
		return nil, 0, fmt.Errorf("semantic search: %w", err)
	}

	if len(result.Results) == 0 {
		slog.Info("typesense semantic search completed (no results)",
			"duration", duration,
		)
		return nil, 0, nil
	}

	item := result.Results[0]
	// Map MultiSearchResultItem fields into api.SearchResult for unified processing.
	sr := &api.SearchResult{
		Found: item.Found,
		Hits:  item.Hits,
	}

	results, total, mapErr := mapSearchResults(sr)

	slog.Info("typesense semantic search completed",
		"duration", duration,
		"total_found", total,
		"returned", len(results),
	)

	return results, total, mapErr
}

// buildFilterBy creates a Typesense filter_by expression that enforces
// visibility rules and applies optional filters for tags, path, date range, etc.
func buildFilterBy(userID string, filters *domain.SearchFilters) string {
	var conditions []string

	// Visibility filter with security
	// - public: anyone can see
	// - private: only the owner can see (MUST enforce user_id)
	if filters != nil && filters.Visibility == "public" {
		conditions = append(conditions, "visibility:=public")
	} else if filters != nil && filters.Visibility == "private" {
		// Private entries: user MUST be authenticated and can only see their own
		if userID == "" {
			// Unauthenticated user trying to see private entries - return impossible condition
			conditions = append(conditions, "visibility:=__impossible__")
		} else {
			conditions = append(conditions, fmt.Sprintf("(visibility:=private && user_id:=%s)", userID))
		}
	} else if userID == "" {
		// Default for unauthenticated: only public
		conditions = append(conditions, "visibility:=public")
	} else {
		// Default for authenticated: public OR own entries
		conditions = append(conditions, fmt.Sprintf("(visibility:=public || user_id:=%s)", userID))
	}

	if filters == nil {
		return strings.Join(conditions, " && ")
	}

	// OnlyMine: restrict to user's own entries only
	if filters.OnlyMine && userID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id:=%s", userID))
	}

	// Tags filter (OR logic)
	if len(filters.Tags) > 0 {
		escapedTags := make([]string, len(filters.Tags))
		for i, tag := range filters.Tags {
			escapedTags[i] = escapeFilterValue(tag)
		}
		conditions = append(conditions, fmt.Sprintf("tags:=[%s]", strings.Join(escapedTags, ", ")))
	}

	// Path filter (exact match)
	if filters.Path != "" {
		conditions = append(conditions, fmt.Sprintf("path:=%s", escapeFilterValue(filters.Path)))
	}

	// Date range filters
	if !filters.FromDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at:>=%d", filters.FromDate.Unix()))
	}
	if !filters.ToDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at:<=%d", filters.ToDate.Unix()))
	}

	return strings.Join(conditions, " && ")
}

// escapeFilterValue escapes special characters in filter values for Typesense.
func escapeFilterValue(val string) string {
	val = strings.ReplaceAll(val, "\\", "\\\\")
	val = strings.ReplaceAll(val, "`", "\\`")
	return fmt.Sprintf("`%s`", val)
}

const contentExcerptLen = 300

// mapSearchResults converts a Typesense search response to domain search results.
func mapSearchResults(result *api.SearchResult) ([]*domain.SearchResult, int, error) {
	if result.Hits == nil {
		return nil, 0, nil
	}

	total := 0
	if result.Found != nil {
		total = *result.Found
	}

	results := make([]*domain.SearchResult, 0, len(*result.Hits))
	for _, hit := range *result.Hits {
		doc := *hit.Document
		s3Key := stringFromDoc(doc, "s3_key")
		sr := &domain.SearchResult{
			EntryID:    stringFromDoc(doc, "entry_id"),
			Title:      stringFromDoc(doc, "title"),
			Summary:    stringFromDoc(doc, "summary"),
			Path:       stringFromDoc(doc, "path"),
			Visibility: stringFromDoc(doc, "visibility"),
			Tags:       stringsFromDoc(doc, "tags"),
			FileType:   stringFromDoc(doc, "file_type"),
			FileSize:   int64FromDoc(doc, "file_size"),
			S3Key:      s3Key,
		}

		// Return inline content only when the entry is NOT stored in S3.
		// Binary/S3-backed entries expose only their summary and file metadata.
		if s3Key == "" {
			if raw := stringFromDoc(doc, "content"); raw != "" {
				if len(raw) <= contentExcerptLen {
					sr.Content = raw
				} else {
					sr.Content = raw[:contentExcerptLen]
				}
			}
		}

		// Convert Unix timestamp to RFC3339 string.
		if ts := int64FromDoc(doc, "created_at"); ts > 0 {
			sr.CreatedAt = time.Unix(ts, 0).UTC().Format(time.RFC3339)
		}

		if hit.VectorDistance != nil {
			// Convert distance to similarity score (lower distance = higher relevance).
			sr.Score = float64(1.0 - *hit.VectorDistance)
		}

		// Collect field-attributed highlights (keyword search only; semantic search
		// does not request HighlightFields so this block is a no-op for it).
		if hit.Highlights != nil {
			for _, h := range *hit.Highlights {
				if h.Field == nil || h.Snippet == nil {
					continue
				}
				sr.Highlights = append(sr.Highlights, domain.Highlight{
					Field:   *h.Field,
					Snippet: *h.Snippet,
				})
			}
		}

		results = append(results, sr)
	}

	return results, total, nil
}

// stringFromDoc safely extracts a string value from a document map.
func stringFromDoc(doc map[string]interface{}, key string) string {
	v, ok := doc[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// stringsFromDoc safely extracts a []string value from a document map.
// Typesense returns string arrays as []interface{} with string elements.
func stringsFromDoc(doc map[string]interface{}, key string) []string {
	v, ok := doc[key]
	if !ok {
		return nil
	}
	raw, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, elem := range raw {
		if s, ok := elem.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// int64FromDoc safely extracts an int64 value from a document map.
// Typesense returns numbers as float64 in JSON.
func int64FromDoc(doc map[string]interface{}, key string) int64 {
	v, ok := doc[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	}
	return 0
}

// floatsToString converts a float32 slice to a comma-separated string for vector queries.
func floatsToString(fs []float32) string {
	if len(fs) == 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(len(fs) * 12) // pre-allocate: ~12 chars per float
	b.WriteString(strconv.FormatFloat(float64(fs[0]), 'f', -1, 32))
	for _, f := range fs[1:] {
		b.WriteByte(',')
		b.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 32))
	}
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
