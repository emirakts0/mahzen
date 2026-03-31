package meilisearch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	meili "github.com/meilisearch/meilisearch-go"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/infra/searchutil"
)

// Indexer implements domain.Indexer using Meilisearch.
type Indexer struct {
	client meili.IndexManager
}

// NewIndexer creates a new Meilisearch-backed indexer.
func NewIndexer(client meili.ServiceManager) *Indexer {
	return &Indexer{client: client.Index(IndexName)}
}

func (idx *Indexer) IndexEntry(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
	doc := buildDocument(entry, tags, embedding)

	slog.Info("meilisearch indexing entry",
		"entry_id", entry.ID,
		"title", entry.Title,
		"path", entry.Path,
		"visibility", entry.Visibility.String(),
		"tag_count", len(tags),
		"tags", doc.Tags,
		"has_embedding", len(embedding) > 0,
	)

	start := time.Now()
	if _, err := idx.client.AddDocumentsWithContext(ctx, []entryDocument{doc}, nil); err != nil {
		slog.Error("meilisearch index entry failed",
			"entry_id", entry.ID,
			"duration", time.Since(start),
			"error", err,
		)
		return fmt.Errorf("indexing entry %s: %w", entry.ID, err)
	}

	slog.Info("meilisearch index entry completed",
		"entry_id", entry.ID,
		"duration", time.Since(start),
	)
	return nil
}

func (idx *Indexer) UpdateEntry(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
	doc := buildDocument(entry, tags, embedding)

	slog.Info("meilisearch updating entry",
		"entry_id", entry.ID,
		"title", entry.Title,
		"has_embedding", len(embedding) > 0,
	)

	start := time.Now()
	if _, err := idx.client.UpdateDocumentsWithContext(ctx, []entryDocument{doc}, nil); err != nil {
		slog.Error("meilisearch update entry failed",
			"entry_id", entry.ID,
			"duration", time.Since(start),
			"error", err,
		)
		return fmt.Errorf("updating entry %s in index: %w", entry.ID, err)
	}

	slog.Info("meilisearch update entry completed",
		"entry_id", entry.ID,
		"duration", time.Since(start),
	)
	return nil
}

func (idx *Indexer) DeleteEntry(ctx context.Context, id string) error {
	slog.Info("meilisearch deleting entry", "entry_id", id)

	start := time.Now()
	if _, err := idx.client.DeleteDocumentWithContext(ctx, id, nil); err != nil {
		slog.Error("meilisearch delete entry failed",
			"entry_id", id,
			"duration", time.Since(start),
			"error", err,
		)
		return fmt.Errorf("deleting entry %s from index: %w", id, err)
	}

	slog.Info("meilisearch delete entry completed",
		"entry_id", id,
		"duration", time.Since(start),
	)
	return nil
}

type entryDocument struct {
	ID         string               `json:"id"`
	UserID     string               `json:"user_id"`
	Title      string               `json:"title"`
	Content    string               `json:"content"`
	Summary    string               `json:"summary"`
	Path       string               `json:"path"`
	Visibility string               `json:"visibility"`
	Tags       []string             `json:"tags"`
	CreatedAt  int64                `json:"created_at"`
	FileType   string               `json:"file_type"`
	FileSize   int64                `json:"file_size"`
	Vectors    map[string][]float32 `json:"_vectors"`
}

func buildDocument(entry *domain.Entry, tags []*domain.Tag, embedding []float32) entryDocument {
	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}

	content := ""
	if searchutil.IsTextReadable(entry.FileType) {
		content = entry.Content
	}

	doc := entryDocument{
		ID:         entry.ID,
		UserID:     entry.UserID,
		Title:      entry.Title,
		Content:    content,
		Summary:    entry.Summary,
		Path:       entry.Path,
		Visibility: entry.Visibility.String(),
		Tags:       tagNames,
		CreatedAt:  entry.CreatedAt.Unix(),
		FileType:   entry.FileType,
		FileSize:   entry.FileSize,
	}

	if len(embedding) > 0 {
		doc.Vectors = map[string][]float32{
			"openai": embedding,
		}
	}

	return doc
}
