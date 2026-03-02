package domain

import "context"

// SearchResult represents a single result from a search query.
type SearchResult struct {
	EntryID    string
	Title      string
	Snippet    string
	Score      float64
	Highlights []string
}

// Indexer defines operations for indexing entries in the search engine.
type Indexer interface {
	IndexEntry(ctx context.Context, entry *Entry, tags []*Tag, embedding []float32) error
	UpdateEntry(ctx context.Context, entry *Entry, tags []*Tag, embedding []float32) error
	DeleteEntry(ctx context.Context, id string) error
}

// Searcher defines search operations over indexed entries.
type Searcher interface {
	KeywordSearch(ctx context.Context, query, userID string, limit, offset int) ([]*SearchResult, int, error)
	SemanticSearch(ctx context.Context, embedding []float32, userID string, limit, offset int) ([]*SearchResult, int, error)
}
