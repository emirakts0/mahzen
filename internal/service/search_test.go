package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service/mock"
)

// ---------------------------------------------------------------------------
// KeywordSearch
// ---------------------------------------------------------------------------

func TestKeywordSearch(t *testing.T) {
	searcher := &mock.Searcher{
		KeywordSearchFn: func(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
			return []*domain.SearchResult{
				{EntryID: "e1", Title: "Go Concurrency", Score: 0.95},
				{EntryID: "e2", Title: "Go Testing", Score: 0.85},
			}, 2, nil
		},
	}

	svc := NewSearchService(searcher, &mock.Embedder{})
	results, total, err := svc.KeywordSearch(context.Background(), "go", "user-1", nil, 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 2, total)
	assert.Equal(t, "Go Concurrency", results[0].Title)
	assert.InDelta(t, 0.95, results[0].Score, 0.001)
}

func TestKeywordSearch_NoResults(t *testing.T) {
	searcher := &mock.Searcher{
		KeywordSearchFn: func(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
			return nil, 0, nil
		},
	}

	svc := NewSearchService(searcher, &mock.Embedder{})
	results, total, err := svc.KeywordSearch(context.Background(), "nonexistent", "user-1", nil, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, 0, total)
}

func TestKeywordSearch_Error(t *testing.T) {
	searcher := &mock.Searcher{
		KeywordSearchFn: func(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
			return nil, 0, errors.New("search unavailable")
		},
	}

	svc := NewSearchService(searcher, &mock.Embedder{})
	_, _, err := svc.KeywordSearch(context.Background(), "go", "user-1", nil, 10, 0)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// SemanticSearch
// ---------------------------------------------------------------------------

func TestSemanticSearch(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3}

	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			assert.Equal(t, "what is concurrency", text)
			return embedding, nil
		},
	}
	searcher := &mock.Searcher{
		SemanticSearchFn: func(ctx context.Context, emb []float32, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
			assert.Equal(t, embedding, emb)
			return []*domain.SearchResult{
				{EntryID: "e1", Title: "Concurrency Patterns", Score: 0.92},
			}, 1, nil
		},
	}

	svc := NewSearchService(searcher, embedder)
	results, total, err := svc.SemanticSearch(context.Background(), "what is concurrency", "user-1", nil, 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, 1, total)
	assert.Equal(t, "Concurrency Patterns", results[0].Title)
}

func TestSemanticSearch_EmbeddingError(t *testing.T) {
	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			return nil, errors.New("openai rate limit")
		},
	}

	svc := NewSearchService(&mock.Searcher{}, embedder)
	_, _, err := svc.SemanticSearch(context.Background(), "query", "user-1", nil, 10, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "embedding query")
}

func TestSemanticSearch_SearcherError(t *testing.T) {
	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			return []float32{0.1}, nil
		},
	}
	searcher := &mock.Searcher{
		SemanticSearchFn: func(ctx context.Context, emb []float32, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
			return nil, 0, errors.New("search failed")
		},
	}

	svc := NewSearchService(searcher, embedder)
	_, _, err := svc.SemanticSearch(context.Background(), "query", "user-1", nil, 10, 0)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// SemanticSearch: verifies the pipeline (embed → search) works correctly
// ---------------------------------------------------------------------------

func TestSemanticSearch_VisibilityFiltering(t *testing.T) {
	// Verifies that userID is correctly passed through to the searcher
	// which is responsible for visibility filtering.
	var receivedUserID string

	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			return []float32{0.5}, nil
		},
	}
	searcher := &mock.Searcher{
		SemanticSearchFn: func(ctx context.Context, emb []float32, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
			receivedUserID = userID
			return nil, 0, nil
		},
	}

	svc := NewSearchService(searcher, embedder)
	_, _, err := svc.SemanticSearch(context.Background(), "test query", "user-42", nil, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, "user-42", receivedUserID)
}

// ---------------------------------------------------------------------------
// SearchFilters
// ---------------------------------------------------------------------------

func TestKeywordSearch_WithFilters(t *testing.T) {
	var receivedFilters *domain.SearchFilters

	searcher := &mock.Searcher{
		KeywordSearchFn: func(ctx context.Context, query, userID string, filters *domain.SearchFilters, limit, offset int) ([]*domain.SearchResult, int, error) {
			receivedFilters = filters
			return []*domain.SearchResult{
				{EntryID: "e1", Title: "Filtered Result", Score: 0.95},
			}, 1, nil
		},
	}

	svc := NewSearchService(searcher, &mock.Embedder{})
	filters := &domain.SearchFilters{
		Tags:     []string{"go", "testing"},
		Path:     "/notes/work",
		OnlyMine: true,
	}
	results, total, err := svc.KeywordSearch(context.Background(), "go", "user-1", filters, 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, 1, total)
	assert.NotNil(t, receivedFilters)
	assert.Equal(t, []string{"go", "testing"}, receivedFilters.Tags)
	assert.Equal(t, "/notes/work", receivedFilters.Path)
	assert.True(t, receivedFilters.OnlyMine)
}
