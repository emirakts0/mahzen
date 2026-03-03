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
func buildDocument(entry *domain.Entry, tags []*domain.Tag, embedding []float32) map[string]interface{} {
	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}

	doc := map[string]interface{}{
		"id":         entry.ID,
		"entry_id":   entry.ID,
		"user_id":    entry.UserID,
		"title":      entry.Title,
		"content":    entry.Content,
		"summary":    entry.Summary,
		"path":       entry.Path,
		"visibility": entry.Visibility.String(),
		"tags":       tagNames,
		"created_at": entry.CreatedAt.Unix(),
	}

	if len(embedding) > 0 {
		doc["embedding"] = embedding
	}

	return doc
}
