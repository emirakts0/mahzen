package typesense

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/typesense/typesense-go/v3/typesense"
	"github.com/typesense/typesense-go/v3/typesense/api"

	"github.com/emirakts0/mahzen/internal/domain"
)

// Indexer implements domain.Indexer using Typesense.
type Indexer struct {
	client *typesense.Client
}

// NewIndexer creates a new Typesense-backed indexer.
func NewIndexer(client *typesense.Client) *Indexer {
	return &Indexer{client: client}
}

func (idx *Indexer) IndexEntry(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
	doc := buildDocument(entry, tags, embedding)

	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}

	slog.Info("typesense indexing entry",
		"entry_id", entry.ID,
		"title", entry.Title,
		"path", entry.Path,
		"visibility", entry.Visibility.String(),
		"tag_count", len(tags),
		"tags", tagNames,
		"has_embedding", len(embedding) > 0,
	)

	start := time.Now()
	if _, err := idx.client.Collection(CollectionName).Documents().Create(ctx, doc, &api.DocumentIndexParameters{}); err != nil {
		slog.Error("typesense index entry failed",
			"entry_id", entry.ID,
			"duration", time.Since(start),
			"error", err,
		)
		return fmt.Errorf("indexing entry %s: %w", entry.ID, err)
	}

	slog.Info("typesense index entry completed",
		"entry_id", entry.ID,
		"duration", time.Since(start),
	)
	return nil
}

func (idx *Indexer) UpdateEntry(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
	doc := buildDocument(entry, tags, embedding)

	slog.Info("typesense updating entry",
		"entry_id", entry.ID,
		"title", entry.Title,
		"has_embedding", len(embedding) > 0,
	)

	start := time.Now()
	if _, err := idx.client.Collection(CollectionName).Document(entry.ID).Update(ctx, doc, &api.DocumentIndexParameters{}); err != nil {
		slog.Error("typesense update entry failed",
			"entry_id", entry.ID,
			"duration", time.Since(start),
			"error", err,
		)
		return fmt.Errorf("updating entry %s in index: %w", entry.ID, err)
	}

	slog.Info("typesense update entry completed",
		"entry_id", entry.ID,
		"duration", time.Since(start),
	)
	return nil
}

func (idx *Indexer) DeleteEntry(ctx context.Context, id string) error {
	slog.Info("typesense deleting entry", "entry_id", id)

	start := time.Now()
	if _, err := idx.client.Collection(CollectionName).Document(id).Delete(ctx); err != nil {
		slog.Error("typesense delete entry failed",
			"entry_id", id,
			"duration", time.Since(start),
			"error", err,
		)
		return fmt.Errorf("deleting entry %s from index: %w", id, err)
	}

	slog.Info("typesense delete entry completed",
		"entry_id", id,
		"duration", time.Since(start),
	)
	return nil
}

// buildDocument converts a domain entry into a Typesense document map.
// For binary file types (e.g. mp4, zip) the content field is omitted from the
// index so only the summary is searchable. Text-readable entries always have
// their full content indexed.
func buildDocument(entry *domain.Entry, tags []*domain.Tag, embedding []float32) map[string]any {
	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}

	doc := map[string]any{
		"id":         entry.ID,
		"entry_id":   entry.ID,
		"user_id":    entry.UserID,
		"title":      entry.Title,
		"summary":    entry.Summary,
		"path":       entry.Path,
		"visibility": entry.Visibility.String(),
		"tags":       tagNames,
		"created_at": entry.CreatedAt.Unix(),
		"file_type":  entry.FileType,
		"file_size":  entry.FileSize,
	}

	// Only index content for text-readable entries.
	// Binary files (mp4, zip, pdf, etc.) should not have their content indexed;
	// only their AI-generated summary will be searchable.
	if isTextReadable(entry.FileType) {
		doc["content"] = entry.Content
	} else {
		doc["content"] = ""
	}

	if len(embedding) > 0 {
		doc["embedding"] = embedding
	}

	return doc
}

// textReadableTypes is the set of file extensions that contain human-readable text
// and whose content should be indexed for full-text search.
var textReadableTypes = map[string]struct{}{
	"":      {}, // plain text entry (no file type)
	"txt":   {},
	"md":    {},
	"go":    {},
	"java":  {},
	"py":    {},
	"js":    {},
	"ts":    {},
	"jsx":   {},
	"tsx":   {},
	"css":   {},
	"html":  {},
	"htm":   {},
	"xml":   {},
	"json":  {},
	"yaml":  {},
	"yml":   {},
	"toml":  {},
	"ini":   {},
	"sh":    {},
	"bash":  {},
	"zsh":   {},
	"rs":    {},
	"c":     {},
	"cpp":   {},
	"h":     {},
	"hpp":   {},
	"cs":    {},
	"rb":    {},
	"php":   {},
	"sql":   {},
	"r":     {},
	"kt":    {},
	"swift": {},
	"scala": {},
	"lua":   {},
	"pl":    {},
	"csv":   {},
	"log":   {},
	"conf":  {},
	"env":   {},
	"tf":    {},
}

// isTextReadable reports whether the given file extension corresponds to a
// human-readable text format that should have its content indexed.
func isTextReadable(fileType string) bool {
	_, ok := textReadableTypes[fileType]
	return ok
}
