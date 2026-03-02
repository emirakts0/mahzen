package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/emirakts0/mahzen/internal/config"
	"github.com/emirakts0/mahzen/internal/domain"
)

// EntryService contains business logic for managing entries.
type EntryService struct {
	entries       domain.EntryRepository
	tags          domain.TagRepository
	storage       domain.ObjectStorage
	indexer       domain.Indexer
	embedder      domain.Embedder
	s3Threshold   int64
}

// NewEntryService creates a new EntryService.
func NewEntryService(
	entries domain.EntryRepository,
	tags domain.TagRepository,
	storage domain.ObjectStorage,
	indexer domain.Indexer,
	embedder domain.Embedder,
	cfg config.EntryConfig,
) *EntryService {
	return &EntryService{
		entries:     entries,
		tags:        tags,
		storage:     storage,
		indexer:     indexer,
		embedder:    embedder,
		s3Threshold: cfg.S3SizeThreshold,
	}
}

// CreateEntry creates a new entry, optionally storing content in S3 and indexing it.
func (s *EntryService) CreateEntry(ctx context.Context, userID, title, content, path string, visibility domain.Visibility, tagIDs []string) (*domain.Entry, error) {
	normalizedPath, err := domain.NormalizePath(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	entry := &domain.Entry{
		UserID:     userID,
		Title:      title,
		Content:    content,
		Path:       normalizedPath,
		Visibility: visibility,
	}

	// Store large content in S3.
	if int64(len(content)) >= s.s3Threshold {
		key := fmt.Sprintf("entries/%s/%s", userID, uuid.New().String())
		if err := s.storage.Upload(ctx, key, strings.NewReader(content), "text/plain", int64(len(content))); err != nil {
			return nil, fmt.Errorf("uploading content to s3: %w", err)
		}
		entry.S3Key = key
		entry.Content = "" // Clear inline content; it's in S3 now.
	}

	if err := s.entries.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("creating entry: %w", err)
	}

	// Attach tags.
	for _, tagID := range tagIDs {
		if err := s.tags.AttachToEntry(ctx, entry.ID, tagID); err != nil {
			slog.Warn("failed to attach tag", "entry_id", entry.ID, "tag_id", tagID, "error", err)
		}
	}

	// Async: generate embedding and index in Typesense.
	go s.indexEntryAsync(entry, tagIDs, content)

	return entry, nil
}

// GetEntry retrieves an entry by ID, fetching content from S3 if needed.
func (s *EntryService) GetEntry(ctx context.Context, id string) (*domain.Entry, error) {
	entry, err := s.entries.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting entry: %w", err)
	}

	if entry.IsStoredInS3() {
		reader, err := s.storage.Download(ctx, entry.S3Key)
		if err != nil {
			return nil, fmt.Errorf("downloading entry content: %w", err)
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("reading entry content: %w", err)
		}
		entry.Content = string(data)
	}

	return entry, nil
}

// UpdateEntry updates an existing entry.
func (s *EntryService) UpdateEntry(ctx context.Context, id, title, content, path string, visibility domain.Visibility, tagIDs []string) (*domain.Entry, error) {
	entry, err := s.entries.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting entry for update: %w", err)
	}

	// Normalize path; if empty, keep the existing path.
	if path != "" {
		normalizedPath, err := domain.NormalizePath(path)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}
		entry.Path = normalizedPath
	}

	entry.Title = title
	entry.Visibility = visibility

	// Handle content storage changes.
	if int64(len(content)) >= s.s3Threshold {
		key := entry.S3Key
		if key == "" {
			key = fmt.Sprintf("entries/%s/%s", entry.UserID, uuid.New().String())
		}
		if err := s.storage.Upload(ctx, key, strings.NewReader(content), "text/plain", int64(len(content))); err != nil {
			return nil, fmt.Errorf("uploading updated content: %w", err)
		}
		entry.S3Key = key
		entry.Content = ""
	} else {
		// Content is small enough for inline storage.
		if entry.S3Key != "" {
			// Clean up old S3 object.
			_ = s.storage.Delete(ctx, entry.S3Key)
			entry.S3Key = ""
		}
		entry.Content = content
	}

	if err := s.entries.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("updating entry: %w", err)
	}

	// Re-sync tags: detach all existing, then attach new ones.
	existingTags, err := s.tags.ListByEntry(ctx, id)
	if err == nil {
		for _, t := range existingTags {
			_ = s.tags.DetachFromEntry(ctx, id, t.ID)
		}
	}
	for _, tagID := range tagIDs {
		if err := s.tags.AttachToEntry(ctx, id, tagID); err != nil {
			slog.Warn("failed to attach tag on update", "entry_id", id, "tag_id", tagID, "error", err)
		}
	}

	go s.indexEntryAsync(entry, tagIDs, content)

	return entry, nil
}

// DeleteEntry deletes an entry and its associated S3 content and search index.
func (s *EntryService) DeleteEntry(ctx context.Context, id string) error {
	entry, err := s.entries.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("getting entry for deletion: %w", err)
	}

	if entry.S3Key != "" {
		if err := s.storage.Delete(ctx, entry.S3Key); err != nil {
			slog.Warn("failed to delete s3 object", "key", entry.S3Key, "error", err)
		}
	}

	if err := s.entries.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting entry: %w", err)
	}

	if err := s.indexer.DeleteEntry(ctx, id); err != nil {
		slog.Warn("failed to delete entry from index", "entry_id", id, "error", err)
	}

	return nil
}

// ListEntries lists entries accessible to the given user, optionally filtered by path prefix.
func (s *EntryService) ListEntries(ctx context.Context, userID, pathPrefix string, limit, offset int) ([]*domain.Entry, int, error) {
	return s.entries.ListAccessible(ctx, userID, pathPrefix, limit, offset)
}

// indexEntryAsync generates an embedding and indexes the entry in Typesense.
// Runs in a background goroutine — errors are logged, not returned.
func (s *EntryService) indexEntryAsync(entry *domain.Entry, tagIDs []string, content string) {
	ctx := context.Background()

	// Resolve tag objects.
	tags := make([]*domain.Tag, 0, len(tagIDs))
	for _, id := range tagIDs {
		tag, err := s.tags.GetByID(ctx, id)
		if err != nil {
			slog.Warn("failed to get tag for indexing", "tag_id", id, "error", err)
			continue
		}
		tags = append(tags, tag)
	}

	// Use content for embedding; fall back to title + summary.
	embedText := content
	if embedText == "" {
		embedText = entry.Title
		if entry.Summary != "" {
			embedText += " " + entry.Summary
		}
	}

	embedding, err := s.embedder.Embed(ctx, embedText)
	if err != nil {
		slog.Error("failed to generate embedding", "entry_id", entry.ID, "error", err)
		embedding = nil // Index without embedding.
	}

	if err := s.indexer.IndexEntry(ctx, entry, tags, embedding); err != nil {
		slog.Error("failed to index entry", "entry_id", entry.ID, "error", err)
	}
}
