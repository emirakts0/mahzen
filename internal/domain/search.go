package domain

import (
	"context"
	"time"
)

// SearchFilters contains optional filters for search queries.
type SearchFilters struct {
	Tags       []string  // Filter by tags (OR logic)
	Path       string    // Filter by exact path match
	FromDate   time.Time // Filter entries created on or after this date
	ToDate     time.Time // Filter entries created on or before this date
	OnlyMine   bool      // When true, only return entries owned by the user
	Visibility string    // "public", "private", or empty for default behavior
}

// Highlight represents a single field-attributed highlight snippet from Typesense.
// Snippet contains the matched text with query tokens.
type Highlight struct {
	Field   string // "title" | "content" | "summary"
	Snippet string
}

// SearchResult represents a single result from a search query.
type SearchResult struct {
	EntryID    string
	Title      string
	Summary    string // AI-generated summary stored in the index.
	Content    string // Inline content from Typesense index; empty when entry is binary or S3-stored.
	Score      float64
	Highlights []Highlight // field-attributed; only populated for keyword search
	Path       string
	Visibility string
	Tags       []string
	CreatedAt  string
	FileType   string // Empty for plain-text entries; extension (e.g. "mp4") for binary files.
	FileSize   int64  // Size of the file in bytes.
	S3Key      string // Set for S3-stored entries; used to generate download URLs.
}

// Indexer defines operations for indexing entries in the search engine.
type Indexer interface {
	IndexEntry(ctx context.Context, entry *Entry, tags []*Tag, embedding []float32) error
	UpdateEntry(ctx context.Context, entry *Entry, tags []*Tag, embedding []float32) error
	DeleteEntry(ctx context.Context, id string) error
}

// Searcher defines search operations over indexed entries.
type Searcher interface {
	KeywordSearch(ctx context.Context, query, userID string, filters *SearchFilters, limit, offset int) ([]*SearchResult, int, error)
	SemanticSearch(ctx context.Context, embedding []float32, userID string, filters *SearchFilters, limit, offset int) ([]*SearchResult, int, error)
}
