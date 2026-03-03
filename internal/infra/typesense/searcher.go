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

func (s *Searcher) KeywordSearch(ctx context.Context, query, userID string, limit, offset int) ([]*domain.SearchResult, int, error) {
	filterBy := buildVisibilityFilter(userID)

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

func (s *Searcher) SemanticSearch(ctx context.Context, embedding []float32, userID string, limit, offset int) ([]*domain.SearchResult, int, error) {
	filterBy := buildVisibilityFilter(userID)
	vectorQuery := fmt.Sprintf("embedding:([%s], k:%d)", floatsToString(embedding), limit)

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

// buildVisibilityFilter creates a Typesense filter_by expression that enforces
// visibility rules: public entries OR entries owned by the requesting user.
func buildVisibilityFilter(userID string) string {
	if userID == "" {
		return "visibility:=public"
	}
	return fmt.Sprintf("visibility:=public || user_id:=%s", userID)
}

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
		sr := &domain.SearchResult{
			EntryID: stringFromDoc(doc, "entry_id"),
			Title:   stringFromDoc(doc, "title"),
			Snippet: stringFromDoc(doc, "summary"),
		}

		if hit.TextMatch != nil {
			// Send raw TextMatch score; normalization is done on the frontend.
			sr.Score = float64(*hit.TextMatch)
		}
		if hit.VectorDistance != nil {
			// Convert distance to similarity score (lower distance = higher relevance)
			sr.Score = float64(1.0 - *hit.VectorDistance)
		}

		if hit.Highlights != nil {
			for _, h := range *hit.Highlights {
				if h.Snippet != nil {
					sr.Highlights = append(sr.Highlights, *h.Snippet)
				}
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
