package typesense

import (
	"context"
	"fmt"

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

	params := &api.SearchCollectionParams{
		Q:              pointer.String(query),
		QueryBy:        pointer.String("title,content,summary,tags"),
		FilterBy:       pointer.String(filterBy),
		PerPage:        pointer.Int(limit),
		Page:           pointer.Int((offset / max(limit, 1)) + 1),
		HighlightFields: pointer.String("title,content,summary"),
	}

	result, err := s.client.Collection(CollectionName).Documents().Search(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("keyword search: %w", err)
	}

	return mapSearchResults(result)
}

func (s *Searcher) SemanticSearch(ctx context.Context, embedding []float32, userID string, limit, offset int) ([]*domain.SearchResult, int, error) {
	filterBy := buildVisibilityFilter(userID)

	// Build the vector query string: embedding field, top-k results
	vectorQuery := fmt.Sprintf("embedding:([%s], k:%d)", floatsToString(embedding), limit)

	params := &api.SearchCollectionParams{
		Q:           pointer.String("*"),
		VectorQuery: pointer.String(vectorQuery),
		FilterBy:    pointer.String(filterBy),
		PerPage:     pointer.Int(limit),
		Page:        pointer.Int((offset / max(limit, 1)) + 1),
	}

	result, err := s.client.Collection(CollectionName).Documents().Search(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("semantic search: %w", err)
	}

	return mapSearchResults(result)
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
	s := fmt.Sprintf("%f", fs[0])
	for _, f := range fs[1:] {
		s += fmt.Sprintf(",%f", f)
	}
	return s
}
