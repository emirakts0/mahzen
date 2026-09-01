package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emirakts0/mahzen/internal/domain"
)

// SearchService provides keyword and semantic search over entries.
type SearchService struct {
	searcher domain.Searcher
	embedder domain.Embedder
}

func NewSearchService(searcher domain.Searcher, embedder domain.Embedder) *SearchService {
	return &SearchService{searcher: searcher, embedder: embedder}
}

// KeywordSearch performs a text-based search.
func (s *SearchService) KeywordSearch(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
	return s.searcher.KeywordSearch(ctx, query, userID, filters, limit, offset)
}

// SemanticSearch converts the query to an embedding and performs vector search.
// Returns no results when no embedding provider is configured.
func (s *SearchService) SemanticSearch(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
	embedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("embedding query: %w", err)
	}
	if len(embedding) == 0 {
		slog.Warn("semantic search skipped: embedding provider not configured")
		return nil, 0, nil
	}

	return s.searcher.SemanticSearch(ctx, embedding, userID, filters, limit, offset)
}
