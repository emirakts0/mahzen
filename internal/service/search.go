package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
func (s *SearchService) KeywordSearch(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
	slog.Info("keyword search request",
		"query", query,
		"user_id", userID,
		"filters", filters,
		"limit", limit,
		"offset", offset,
	)

	start := time.Now()
	results, total, err := s.searcher.KeywordSearch(ctx, query, userID, filters, limit, offset)
	duration := time.Since(start)

	if err != nil {
		slog.Error("keyword search failed",
			"query", query,
			"duration", duration,
			"error", err,
		)
		return nil, 0, err
	}

	slog.Info("keyword search completed",
		"query", query,
		"duration", duration,
		"total", total,
		"returned", len(results),
	)
	return results, total, nil
}

// SemanticSearch converts the query to an embedding and performs vector search.
func (s *SearchService) SemanticSearch(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
	slog.Info("semantic search request",
		"query", query,
		"user_id", userID,
		"filters", filters,
		"limit", limit,
		"offset", offset,
	)

	embedStart := time.Now()
	embedding, err := s.embedder.Embed(ctx, query)
	embedDuration := time.Since(embedStart)

	if err != nil {
		slog.Error("semantic search embedding failed",
			"query", query,
			"duration", embedDuration,
			"error", err,
		)
		return nil, 0, fmt.Errorf("embedding query: %w", err)
	}

	// No embedding provider configured — semantic search unavailable.
	if len(embedding) == 0 {
		slog.Warn("semantic search skipped: embedding provider not configured")
		return nil, 0, nil
	}

	slog.Info("semantic search embedding ready",
		"query", query,
		"embed_duration", embedDuration,
		"dimensions", len(embedding),
	)

	searchStart := time.Now()
	results, total, err := s.searcher.SemanticSearch(ctx, embedding, userID, filters, limit, offset)
	searchDuration := time.Since(searchStart)

	if err != nil {
		slog.Error("semantic search failed",
			"query", query,
			"search_duration", searchDuration,
			"error", err,
		)
		return nil, 0, err
	}

	slog.Info("semantic search completed",
		"query", query,
		"embed_duration", embedDuration,
		"search_duration", searchDuration,
		"total_duration", time.Since(embedStart),
		"total", total,
		"returned", len(results),
	)
	return results, total, nil
}
