package service

import (
	"context"
	"fmt"

	"github.com/emirakts0/mahzen/internal/domain"
)

// SearchService provides keyword and semantic search over entries.
type SearchService struct {
	searcher domain.Searcher
	embedder domain.Embedder
}

// NewSearchService creates a new SearchService.
func NewSearchService(searcher domain.Searcher, embedder domain.Embedder) *SearchService {
	return &SearchService{
		searcher: searcher,
		embedder: embedder,
	}
}

// KeywordSearch performs a text-based search.
func (s *SearchService) KeywordSearch(ctx context.Context, query, userID string, limit, offset int) ([]*domain.SearchResult, int, error) {
	return s.searcher.KeywordSearch(ctx, query, userID, limit, offset)
}

// SemanticSearch converts the query to an embedding and performs vector search.
func (s *SearchService) SemanticSearch(ctx context.Context, query, userID string, limit, offset int) ([]*domain.SearchResult, int, error) {
	embedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("embedding query: %w", err)
	}

	return s.searcher.SemanticSearch(ctx, embedding, userID, limit, offset)
}
