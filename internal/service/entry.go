package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/emirakts0/mahzen/internal/config"
	"github.com/emirakts0/mahzen/internal/domain"
)

// EntryService contains business logic for managing entries.
type EntryService struct {
	entries     domain.EntryRepository
	tags        domain.TagRepository
	storage     domain.ObjectStorage
	indexer     domain.Indexer
	embedder    domain.Embedder
	s3Threshold int64
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
// fileType is the client-provided file extension (e.g. "mp4", "zip"). Empty string means plain text.
func (s *EntryService) CreateEntry(ctx context.Context, userID, title, content, path, fileType string, visibility domain.Visibility, tagIDs []string) (*domain.Entry, error) {
	slog.Info("creating entry",
		"user_id", userID,
		"title", title,
		"path", path,
		"visibility", visibility.String(),
		"content_length", len(content),
		"file_type", fileType,
		"tag_count", len(tagIDs),
	)

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
		FileType:   fileType,
		FileSize:   int64(len(content)),
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
		slog.Error("failed to create entry in db",
			"user_id", userID,
			"title", title,
			"error", err,
		)
		return nil, fmt.Errorf("creating entry: %w", err)
	}

	slog.Info("entry created",
		"entry_id", entry.ID,
		"user_id", userID,
		"title", title,
		"path", normalizedPath,
		"stored_in_s3", entry.IsStoredInS3(),
	)

	// Attach tags.
	for _, tagID := range tagIDs {
		if err := s.tags.AttachToEntry(ctx, entry.ID, tagID); err != nil {
			slog.Warn("failed to attach tag", "entry_id", entry.ID, "tag_id", tagID, "error", err)
		}
	}

	// Async: generate embedding and index in Typesense.
	go s.indexEntryAsync(entry, tagIDs, content, false)

	return entry, nil
}

// GetEntry retrieves an entry by ID, fetching content from S3 if needed.
func (s *EntryService) GetEntry(ctx context.Context, id string) (*domain.Entry, error) {
	slog.Info("getting entry", "entry_id", id)

	entry, err := s.entries.GetByID(ctx, id)
	if err != nil {
		slog.Warn("entry not found", "entry_id", id, "error", err)
		return nil, fmt.Errorf("getting entry: %w", err)
	}

	if entry.IsStoredInS3() {
		slog.Info("fetching entry content from s3",
			"entry_id", id,
			"s3_key", entry.S3Key,
		)
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
// fileType is the client-provided file extension. Empty string keeps the existing value.
func (s *EntryService) UpdateEntry(ctx context.Context, id, title, content, path, fileType string, visibility domain.Visibility, tagIDs []string) (*domain.Entry, error) {
	slog.Info("updating entry",
		"entry_id", id,
		"title", title,
		"path", path,
		"visibility", visibility.String(),
		"content_length", len(content),
		"file_type", fileType,
		"tag_count", len(tagIDs),
	)

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
	if fileType != "" {
		entry.FileType = fileType
	}
	entry.FileSize = int64(len(content))

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
		slog.Error("failed to update entry in db", "entry_id", id, "error", err)
		return nil, fmt.Errorf("updating entry: %w", err)
	}

	slog.Info("entry updated",
		"entry_id", id,
		"title", entry.Title,
		"path", entry.Path,
		"stored_in_s3", entry.IsStoredInS3(),
	)

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

	go s.indexEntryAsync(entry, tagIDs, content, true)

	return entry, nil
}

// DeleteEntry deletes an entry and its associated S3 content and search index.
func (s *EntryService) DeleteEntry(ctx context.Context, id string) error {
	slog.Info("deleting entry", "entry_id", id)

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
		slog.Error("failed to delete entry from db", "entry_id", id, "error", err)
		return fmt.Errorf("deleting entry: %w", err)
	}

	slog.Info("entry deleted", "entry_id", id)

	if err := s.indexer.DeleteEntry(ctx, id); err != nil {
		slog.Warn("failed to delete entry from index", "entry_id", id, "error", err)
	}

	return nil
}

// GetEntryTags returns the tag names attached to the given entry.
func (s *EntryService) GetEntryTags(ctx context.Context, entryID string) ([]string, error) {
	tags, err := s.tags.ListByEntry(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("listing tags for entry: %w", err)
	}
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names, nil
}

// GetEntryTagsBatch fetches tags for multiple entries in a single DB query.
// Returns a map of entryID -> tag names.
func (s *EntryService) GetEntryTagsBatch(ctx context.Context, entryIDs []string) (map[string][]string, error) {
	if len(entryIDs) == 0 {
		return map[string][]string{}, nil
	}
	result, err := s.tags.ListByEntries(ctx, entryIDs)
	if err != nil {
		return nil, fmt.Errorf("batch listing tags for entries: %w", err)
	}
	return result, nil
}

// ListEntries lists entries accessible to the given user, optionally filtered by path prefix.
func (s *EntryService) ListEntries(ctx context.Context, userID, pathPrefix string, limit, offset int) ([]*domain.Entry, int, error) {
	slog.Info("listing entries",
		"user_id", userID,
		"path_prefix", pathPrefix,
		"limit", limit,
		"offset", offset,
	)

	entries, total, err := s.entries.ListAccessible(ctx, userID, pathPrefix, limit, offset)
	if err != nil {
		slog.Error("failed to list entries", "user_id", userID, "error", err)
		return nil, 0, err
	}

	slog.Info("entries listed",
		"user_id", userID,
		"total", total,
		"returned", len(entries),
	)
	return entries, total, nil
}

// indexEntryAsync generates an embedding and indexes the entry in Typesense.
// Runs in a background goroutine — errors are logged, not returned.
// isUpdate controls whether to call UpdateEntry (for updates) or IndexEntry (for creates).
func (s *EntryService) indexEntryAsync(entry *domain.Entry, tagIDs []string, content string, isUpdate bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	slog.Info("async indexing started", "entry_id", entry.ID, "title", entry.Title, "is_update", isUpdate)

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

	slog.Info("generating embedding for entry",
		"entry_id", entry.ID,
		"embed_text_length", len(embedText),
	)

	embedding, err := s.embedder.Embed(ctx, embedText)
	if err != nil {
		slog.Error("failed to generate embedding", "entry_id", entry.ID, "error", err)
		embedding = nil // Index without embedding.
	} else {
		slog.Info("embedding generated",
			"entry_id", entry.ID,
			"dimensions", len(embedding),
		)
	}

	indexedEntry := *entry
	if content != "" {
		indexedEntry.Content = content
	}

	var indexErr error
	if isUpdate {
		indexErr = s.indexer.UpdateEntry(ctx, &indexedEntry, tags, embedding)
	} else {
		indexErr = s.indexer.IndexEntry(ctx, &indexedEntry, tags, embedding)
	}
	if indexErr != nil {
		slog.Error("failed to index entry", "entry_id", entry.ID, "is_update", isUpdate, "error", indexErr)
	} else {
		slog.Info("async indexing completed", "entry_id", entry.ID)
	}
}
