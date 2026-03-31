package service

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/emirakts0/mahzen/internal/domain"
)

// embeddingTimeout is the timeout for embedding generation operations.
const embeddingTimeout = 2 * time.Minute

// FolderInfo represents information about a folder in the entry tree.
type FolderInfo struct {
	Path  string
	Count int
}

// EntryService contains business logic for managing entries.
type EntryService struct {
	entries  domain.EntryRepository
	tags     domain.TagRepository
	indexer  domain.Indexer
	embedder domain.Embedder
}

// NewEntryService creates a new EntryService.
func NewEntryService(
	entries domain.EntryRepository,
	tags domain.TagRepository,
	indexer domain.Indexer,
	embedder domain.Embedder,
) *EntryService {
	return &EntryService{
		entries:  entries,
		tags:     tags,
		indexer:  indexer,
		embedder: embedder,
	}
}

// CreateEntry creates a new entry and indexes it.
// fileType is the client-provided file extension (e.g. "md", "txt"). Empty string means plain text.
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
	)

	// Attach tags.
	for _, tagID := range tagIDs {
		if err := s.tags.AttachToEntry(ctx, entry.ID, tagID); err != nil {
			slog.Warn("failed to attach tag", "entry_id", entry.ID, "tag_id", tagID, "error", err)
		}
	}

	// Async: generate embedding and index in the search engine.
	go s.indexEntryAsync(entry, tagIDs, content, false)

	return entry, nil
}

// GetEntry retrieves an entry by ID.
func (s *EntryService) GetEntry(ctx context.Context, id string) (*domain.Entry, error) {
	slog.Info("getting entry", "entry_id", id)

	entry, err := s.entries.GetByID(ctx, id)
	if err != nil {
		slog.Warn("entry not found", "entry_id", id, "error", err)
		return nil, fmt.Errorf("getting entry: %w", err)
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
	entry.Content = content
	entry.Visibility = visibility
	if fileType != "" {
		entry.FileType = fileType
	}
	entry.FileSize = int64(len(content))

	if err := s.entries.Update(ctx, entry); err != nil {
		slog.Error("failed to update entry in db", "entry_id", id, "error", err)
		return nil, fmt.Errorf("updating entry: %w", err)
	}

	slog.Info("entry updated",
		"entry_id", id,
		"title", entry.Title,
		"path", entry.Path,
	)

	// Re-sync tags: detach all existing, then attach new ones.
	existingTags, err := s.tags.ListByEntry(ctx, id)
	if err == nil {
		for _, t := range existingTags {
			if err := s.tags.DetachFromEntry(ctx, id, t.ID); err != nil {
				slog.Warn("failed to detach tag during update", "entry_id", id, "tag_id", t.ID, "error", err)
			}
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

// DeleteEntry deletes an entry and its search index.
func (s *EntryService) DeleteEntry(ctx context.Context, id string) error {
	slog.Info("deleting entry", "entry_id", id)

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

// indexEntryAsync generates an embedding and indexes the entry in the search engine.
// Runs in a background goroutine — errors are logged, not returned.
// isUpdate controls whether to call UpdateEntry (for updates) or IndexEntry (for creates).
func (s *EntryService) indexEntryAsync(entry *domain.Entry, tagIDs []string, content string, isUpdate bool) {
	ctx, cancel := context.WithTimeout(context.Background(), embeddingTimeout)
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

	embedText := entry.EmbedText()

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

		// Save embedding to database.
		if err := s.entries.UpdateEmbedding(ctx, entry.ID, embedding); err != nil {
			slog.Error("failed to save embedding to db", "entry_id", entry.ID, "error", err)
		} else {
			slog.Info("embedding saved to db", "entry_id", entry.ID)
		}
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

// ListFolders returns all distinct folder paths accessible to the user.
func (s *EntryService) ListFolders(ctx context.Context, userID string) ([]string, error) {
	slog.Info("listing folders", "user_id", userID)

	paths, err := s.entries.ListDistinctPaths(ctx, userID)
	if err != nil {
		slog.Error("failed to list folders", "user_id", userID, "error", err)
		return nil, fmt.Errorf("listing folders: %w", err)
	}

	folderSet := make(map[string]struct{}, len(paths)*2)
	for _, p := range paths {
		folderSet[p] = struct{}{}
		segments := strings.Split(p, "/")
		for i := 1; i < len(segments); i++ {
			parent := strings.Join(segments[:i], "/")
			if parent == "" {
				parent = "/"
			}
			folderSet[parent] = struct{}{}
		}
	}

	folders := make([]string, 0, len(folderSet))
	for f := range folderSet {
		folders = append(folders, f)
	}

	slices.Sort(folders)

	slog.Info("folders listed", "user_id", userID, "count", len(folders))
	return folders, nil
}

// ListChildren returns entries directly in a path and direct subfolders with counts.
// For path "/abc", returns entries where path="/abc" and folders like "/abc/def" (not recursive).
func (s *EntryService) ListChildren(ctx context.Context, userID, path string, own bool, filter *domain.ListEntriesFilter, limit, offset int) ([]*domain.Entry, []FolderInfo, int, error) {
	slog.Info("listing children",
		"user_id", userID,
		"path", path,
		"own", own,
		"visibility", filter.Visibility,
		"tags", filter.Tags,
		"from_date", filter.FromDate,
		"to_date", filter.ToDate,
		"limit", limit,
		"offset", offset,
	)

	entries, directCount, err := s.entries.ListInPath(ctx, userID, path, own, filter, limit, offset)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("listing entries in path: %w", err)
	}

	pathCounts, err := s.entries.ListPathCountsUnderPrefix(ctx, userID, path, own, filter)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("listing path counts under prefix: %w", err)
	}

	folderInfos, subfolderTotal := extractFolderInfos(path, pathCounts)

	// Total includes entries directly at this path + all entries in subfolders
	total := directCount + subfolderTotal

	slog.Info("children listed", "user_id", userID, "path", path, "own", own, "entries", len(entries), "folders", len(folderInfos), "total", total)
	return entries, folderInfos, total, nil
}

// extractFolderInfos extracts direct subfolder paths with counts from path counts under a prefix.
// For prefix "/abc" and pathCounts [{"/abc/def", 3}, {"/abc/def/sub", 2}, {"/abc/ghi", 1}],
// returns [{Path: "/abc/def", Count: 5}, {Path: "/abc/ghi", Count: 1}], subfolderTotal=6.
func extractFolderInfos(prefix string, pathCounts []domain.PathCount) ([]FolderInfo, int) {
	folderCounts := make(map[string]int)
	var subfolderTotal int

	for _, pc := range pathCounts {
		subfolderTotal += pc.Count

		if prefix == "/" {
			remaining := strings.TrimPrefix(pc.Path, "/")
			if remaining == "" {
				continue
			}
			parts := strings.SplitN(remaining, "/", 2)
			if parts[0] != "" {
				fullPath := "/" + parts[0]
				folderCounts[fullPath] += pc.Count
			}
		} else {
			remaining := strings.TrimPrefix(pc.Path, prefix+"/")
			if remaining == "" || remaining == pc.Path {
				continue
			}
			parts := strings.SplitN(remaining, "/", 2)
			if parts[0] != "" {
				fullPath := prefix + "/" + parts[0]
				folderCounts[fullPath] += pc.Count
			}
		}
	}

	folderInfos := make([]FolderInfo, 0, len(folderCounts))
	for path, count := range folderCounts {
		folderInfos = append(folderInfos, FolderInfo{Path: path, Count: count})
	}

	slices.SortFunc(folderInfos, func(a, b FolderInfo) int {
		switch {
		case a.Path < b.Path:
			return -1
		case a.Path > b.Path:
			return 1
		default:
			return 0
		}
	})

	return folderInfos, subfolderTotal
}
