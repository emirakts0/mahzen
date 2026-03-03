package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/emirakts0/mahzen/internal/config"
	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service/mock"
)

// newTestEntryService constructs an EntryService wired to mocks.
func newTestEntryService(
	entries *mock.EntryRepository,
	tags *mock.TagRepository,
	storage *mock.ObjectStorage,
	indexer *mock.Indexer,
	embedder *mock.Embedder,
	threshold int64,
) *EntryService {
	return NewEntryService(entries, tags, storage, indexer, embedder, config.EntryConfig{
		S3SizeThreshold: threshold,
	})
}

// ---------------------------------------------------------------------------
// CreateEntry
// ---------------------------------------------------------------------------

func TestCreateEntry_InlineContent(t *testing.T) {
	entries := &mock.EntryRepository{
		CreateFn: func(ctx context.Context, entry *domain.Entry) error {
			entry.ID = "entry-1"
			entry.CreatedAt = time.Now()
			return nil
		},
	}
	tags := &mock.TagRepository{
		AttachToEntryFn: func(ctx context.Context, entryID, tagID string) error {
			return nil
		},
		GetByIDFn: func(ctx context.Context, id string) (*domain.Tag, error) {
			return &domain.Tag{ID: id, Name: "go", Slug: "go"}, nil
		},
	}
	storage := &mock.ObjectStorage{}
	indexer := &mock.Indexer{
		IndexEntryFn: func(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
			return nil
		},
	}
	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			return []float32{0.1, 0.2, 0.3}, nil
		},
	}

	svc := newTestEntryService(entries, tags, storage, indexer, embedder, 65536)

	entry, err := svc.CreateEntry(context.Background(), "user-1", "Test Title", "short content", "/notes", domain.VisibilityPublic, []string{"tag-1"})
	require.NoError(t, err)
	assert.Equal(t, "entry-1", entry.ID)
	assert.Equal(t, "Test Title", entry.Title)
	assert.Equal(t, "short content", entry.Content)
	assert.Equal(t, "/notes", entry.Path)
	assert.Empty(t, entry.S3Key, "inline content should not have S3 key")
	assert.Equal(t, domain.VisibilityPublic, entry.Visibility)
}

func TestCreateEntry_S3Content(t *testing.T) {
	var uploadedKey string
	var uploadedData string

	entries := &mock.EntryRepository{
		CreateFn: func(ctx context.Context, entry *domain.Entry) error {
			entry.ID = "entry-2"
			return nil
		},
	}
	tags := &mock.TagRepository{
		AttachToEntryFn: func(ctx context.Context, entryID, tagID string) error {
			return nil
		},
		GetByIDFn: func(ctx context.Context, id string) (*domain.Tag, error) {
			return &domain.Tag{ID: id, Name: "test"}, nil
		},
	}
	storage := &mock.ObjectStorage{
		UploadFn: func(ctx context.Context, key string, reader io.Reader, contentType string, size int64) error {
			uploadedKey = key
			data, _ := io.ReadAll(reader)
			uploadedData = string(data)
			return nil
		},
	}
	indexer := &mock.Indexer{
		IndexEntryFn: func(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
			return nil
		},
	}
	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			return []float32{0.1}, nil
		},
	}

	// Threshold of 10 bytes; content is larger.
	svc := newTestEntryService(entries, tags, storage, indexer, embedder, 10)
	largeContent := "this content is well above 10 bytes threshold"

	entry, err := svc.CreateEntry(context.Background(), "user-1", "Big Entry", largeContent, "/docs", domain.VisibilityPrivate, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, uploadedKey, "should have uploaded to S3")
	assert.Equal(t, largeContent, uploadedData)
	assert.Equal(t, uploadedKey, entry.S3Key)
	assert.Empty(t, entry.Content, "inline content should be cleared when stored in S3")
}

func TestCreateEntry_RepoError(t *testing.T) {
	entries := &mock.EntryRepository{
		CreateFn: func(ctx context.Context, entry *domain.Entry) error {
			return errors.New("db connection lost")
		},
	}
	tags := &mock.TagRepository{}
	storage := &mock.ObjectStorage{}
	indexer := &mock.Indexer{}
	embedder := &mock.Embedder{}

	svc := newTestEntryService(entries, tags, storage, indexer, embedder, 65536)
	_, err := svc.CreateEntry(context.Background(), "user-1", "Title", "content", "/", domain.VisibilityPublic, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating entry")
}

func TestCreateEntry_S3UploadError(t *testing.T) {
	entries := &mock.EntryRepository{}
	tags := &mock.TagRepository{}
	storage := &mock.ObjectStorage{
		UploadFn: func(ctx context.Context, key string, reader io.Reader, contentType string, size int64) error {
			return errors.New("s3 unavailable")
		},
	}
	indexer := &mock.Indexer{}
	embedder := &mock.Embedder{}

	svc := newTestEntryService(entries, tags, storage, indexer, embedder, 10)
	_, err := svc.CreateEntry(context.Background(), "user-1", "Title", "large enough content", "/", domain.VisibilityPublic, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "uploading content to s3")
}

func TestCreateEntry_TagAttachError_NonFatal(t *testing.T) {
	entries := &mock.EntryRepository{
		CreateFn: func(ctx context.Context, entry *domain.Entry) error {
			entry.ID = "entry-3"
			return nil
		},
	}
	tags := &mock.TagRepository{
		AttachToEntryFn: func(ctx context.Context, entryID, tagID string) error {
			return errors.New("tag not found")
		},
		GetByIDFn: func(ctx context.Context, id string) (*domain.Tag, error) {
			return nil, errors.New("tag not found")
		},
	}
	storage := &mock.ObjectStorage{}
	indexer := &mock.Indexer{
		IndexEntryFn: func(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
			return nil
		},
	}
	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			return nil, nil
		},
	}

	svc := newTestEntryService(entries, tags, storage, indexer, embedder, 65536)
	// Tag attach fails but CreateEntry should still succeed.
	entry, err := svc.CreateEntry(context.Background(), "user-1", "Title", "content", "/", domain.VisibilityPublic, []string{"bad-tag"})
	require.NoError(t, err)
	assert.Equal(t, "entry-3", entry.ID)
}

// ---------------------------------------------------------------------------
// GetEntry
// ---------------------------------------------------------------------------

func TestGetEntry_InlineContent(t *testing.T) {
	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{
				ID:      id,
				Title:   "My Entry",
				Content: "inline content",
			}, nil
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, &mock.ObjectStorage{}, &mock.Indexer{}, &mock.Embedder{}, 65536)

	entry, err := svc.GetEntry(context.Background(), "entry-1")
	require.NoError(t, err)
	assert.Equal(t, "inline content", entry.Content)
}

func TestGetEntry_S3Content(t *testing.T) {
	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{
				ID:    id,
				Title: "S3 Entry",
				S3Key: "entries/user-1/abc",
			}, nil
		},
	}
	storage := &mock.ObjectStorage{
		DownloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("content from s3")), nil
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, storage, &mock.Indexer{}, &mock.Embedder{}, 65536)

	entry, err := svc.GetEntry(context.Background(), "entry-1")
	require.NoError(t, err)
	assert.Equal(t, "content from s3", entry.Content)
}

func TestGetEntry_NotFound(t *testing.T) {
	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, &mock.ObjectStorage{}, &mock.Indexer{}, &mock.Embedder{}, 65536)
	_, err := svc.GetEntry(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting entry")
}

func TestGetEntry_S3DownloadError(t *testing.T) {
	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{ID: id, S3Key: "entries/user-1/abc"}, nil
		},
	}
	storage := &mock.ObjectStorage{
		DownloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return nil, errors.New("s3 error")
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, storage, &mock.Indexer{}, &mock.Embedder{}, 65536)
	_, err := svc.GetEntry(context.Background(), "entry-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "downloading entry content")
}

// ---------------------------------------------------------------------------
// UpdateEntry
// ---------------------------------------------------------------------------

func TestUpdateEntry_InlineToInline(t *testing.T) {
	var updatedEntry *domain.Entry

	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{ID: id, UserID: "user-1", Content: "old content"}, nil
		},
		UpdateFn: func(ctx context.Context, entry *domain.Entry) error {
			updatedEntry = entry
			return nil
		},
	}
	tags := &mock.TagRepository{
		ListByEntryFn: func(ctx context.Context, entryID string) ([]*domain.Tag, error) {
			return nil, nil
		},
		AttachToEntryFn: func(ctx context.Context, entryID, tagID string) error {
			return nil
		},
		GetByIDFn: func(ctx context.Context, id string) (*domain.Tag, error) {
			return &domain.Tag{ID: id, Name: "test"}, nil
		},
	}
	indexer := &mock.Indexer{
		UpdateEntryFn: func(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
			return nil
		},
	}
	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			return nil, nil
		},
	}

	svc := newTestEntryService(entries, tags, &mock.ObjectStorage{}, indexer, embedder, 65536)

	entry, err := svc.UpdateEntry(context.Background(), "entry-1", "Updated Title", "new content", "/projects", domain.VisibilityPrivate, nil)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", entry.Title)
	assert.Equal(t, "new content", updatedEntry.Content)
	assert.Equal(t, domain.VisibilityPrivate, entry.Visibility)
}

func TestUpdateEntry_InlineToS3(t *testing.T) {
	var uploadCalled bool

	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{ID: id, UserID: "user-1", Content: "small"}, nil
		},
		UpdateFn: func(ctx context.Context, entry *domain.Entry) error {
			return nil
		},
	}
	tags := &mock.TagRepository{
		ListByEntryFn: func(ctx context.Context, entryID string) ([]*domain.Tag, error) {
			return nil, nil
		},
		GetByIDFn: func(ctx context.Context, id string) (*domain.Tag, error) {
			return &domain.Tag{ID: id}, nil
		},
	}
	storage := &mock.ObjectStorage{
		UploadFn: func(ctx context.Context, key string, reader io.Reader, contentType string, size int64) error {
			uploadCalled = true
			return nil
		},
	}
	indexer := &mock.Indexer{
		UpdateEntryFn: func(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
			return nil
		},
	}
	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			return nil, nil
		},
	}

	svc := newTestEntryService(entries, tags, storage, indexer, embedder, 10)

	entry, err := svc.UpdateEntry(context.Background(), "entry-1", "Title", "this is now large enough for s3", "", domain.VisibilityPublic, nil)
	require.NoError(t, err)
	assert.True(t, uploadCalled, "should upload to S3")
	assert.NotEmpty(t, entry.S3Key)
	assert.Empty(t, entry.Content, "inline content should be cleared")
}

func TestUpdateEntry_S3ToInline(t *testing.T) {
	var deletedKey string

	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{ID: id, UserID: "user-1", S3Key: "entries/user-1/old-key"}, nil
		},
		UpdateFn: func(ctx context.Context, entry *domain.Entry) error {
			return nil
		},
	}
	tags := &mock.TagRepository{
		ListByEntryFn: func(ctx context.Context, entryID string) ([]*domain.Tag, error) {
			return nil, nil
		},
		GetByIDFn: func(ctx context.Context, id string) (*domain.Tag, error) {
			return &domain.Tag{ID: id}, nil
		},
	}
	storage := &mock.ObjectStorage{
		DeleteFn: func(ctx context.Context, key string) error {
			deletedKey = key
			return nil
		},
	}
	indexer := &mock.Indexer{
		UpdateEntryFn: func(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
			return nil
		},
	}
	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			return nil, nil
		},
	}

	svc := newTestEntryService(entries, tags, storage, indexer, embedder, 65536)

	entry, err := svc.UpdateEntry(context.Background(), "entry-1", "Title", "small now", "", domain.VisibilityPublic, nil)
	require.NoError(t, err)
	assert.Equal(t, "entries/user-1/old-key", deletedKey, "old S3 object should be deleted")
	assert.Empty(t, entry.S3Key, "S3 key should be cleared")
	assert.Equal(t, "small now", entry.Content)
}

func TestUpdateEntry_ReTagging(t *testing.T) {
	var detachedTags []string
	var attachedTags []string

	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{ID: id, UserID: "user-1", Content: "content"}, nil
		},
		UpdateFn: func(ctx context.Context, entry *domain.Entry) error {
			return nil
		},
	}
	tags := &mock.TagRepository{
		ListByEntryFn: func(ctx context.Context, entryID string) ([]*domain.Tag, error) {
			return []*domain.Tag{
				{ID: "old-tag-1", Name: "old"},
			}, nil
		},
		DetachFromEntryFn: func(ctx context.Context, entryID, tagID string) error {
			detachedTags = append(detachedTags, tagID)
			return nil
		},
		AttachToEntryFn: func(ctx context.Context, entryID, tagID string) error {
			attachedTags = append(attachedTags, tagID)
			return nil
		},
		GetByIDFn: func(ctx context.Context, id string) (*domain.Tag, error) {
			return &domain.Tag{ID: id, Name: "new"}, nil
		},
	}
	indexer := &mock.Indexer{
		UpdateEntryFn: func(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
			return nil
		},
	}
	embedder := &mock.Embedder{
		EmbedFn: func(ctx context.Context, text string) ([]float32, error) {
			return nil, nil
		},
	}

	svc := newTestEntryService(entries, tags, &mock.ObjectStorage{}, indexer, embedder, 65536)
	_, err := svc.UpdateEntry(context.Background(), "entry-1", "Title", "content", "", domain.VisibilityPublic, []string{"new-tag-1", "new-tag-2"})
	require.NoError(t, err)
	assert.Equal(t, []string{"old-tag-1"}, detachedTags, "old tags should be detached")
	assert.Equal(t, []string{"new-tag-1", "new-tag-2"}, attachedTags, "new tags should be attached")
}

// ---------------------------------------------------------------------------
// DeleteEntry
// ---------------------------------------------------------------------------

func TestDeleteEntry_WithS3(t *testing.T) {
	var s3Deleted, indexDeleted bool

	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{ID: id, S3Key: "entries/user-1/abc"}, nil
		},
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	storage := &mock.ObjectStorage{
		DeleteFn: func(ctx context.Context, key string) error {
			s3Deleted = true
			return nil
		},
	}
	indexer := &mock.Indexer{
		DeleteEntryFn: func(ctx context.Context, id string) error {
			indexDeleted = true
			return nil
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, storage, indexer, &mock.Embedder{}, 65536)
	err := svc.DeleteEntry(context.Background(), "entry-1")
	require.NoError(t, err)
	assert.True(t, s3Deleted, "S3 object should be deleted")
	assert.True(t, indexDeleted, "search index entry should be deleted")
}

func TestDeleteEntry_WithoutS3(t *testing.T) {
	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{ID: id, Content: "inline"}, nil
		},
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	indexer := &mock.Indexer{
		DeleteEntryFn: func(ctx context.Context, id string) error {
			return nil
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, &mock.ObjectStorage{}, indexer, &mock.Embedder{}, 65536)
	err := svc.DeleteEntry(context.Background(), "entry-1")
	require.NoError(t, err)
}

func TestDeleteEntry_NotFound(t *testing.T) {
	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, &mock.ObjectStorage{}, &mock.Indexer{}, &mock.Embedder{}, 65536)
	err := svc.DeleteEntry(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting entry for deletion")
}

func TestDeleteEntry_S3DeleteError_NonFatal(t *testing.T) {
	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{ID: id, S3Key: "entries/user-1/abc"}, nil
		},
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	storage := &mock.ObjectStorage{
		DeleteFn: func(ctx context.Context, key string) error {
			return errors.New("s3 timeout")
		},
	}
	indexer := &mock.Indexer{
		DeleteEntryFn: func(ctx context.Context, id string) error {
			return nil
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, storage, indexer, &mock.Embedder{}, 65536)
	// S3 delete fails but entry deletion should still succeed
	// because of the error check (err != nil but it continues only via slog.Warn).
	// Wait - actually looking at the code, if S3 delete fails, it returns an error!
	// Let me re-read the code...
	// Lines 166-168: if err := s.storage.Delete(...); err != nil { slog.Warn(...) }
	// It logs a warning but does NOT return. Good, so it's non-fatal.
	err := svc.DeleteEntry(context.Background(), "entry-1")
	require.NoError(t, err, "S3 delete failure should be non-fatal")
}

func TestDeleteEntry_RepoDeleteError(t *testing.T) {
	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{ID: id}, nil
		},
		DeleteFn: func(ctx context.Context, id string) error {
			return errors.New("constraint violation")
		},
	}
	indexer := &mock.Indexer{}

	svc := newTestEntryService(entries, &mock.TagRepository{}, &mock.ObjectStorage{}, indexer, &mock.Embedder{}, 65536)
	err := svc.DeleteEntry(context.Background(), "entry-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleting entry")
}

// ---------------------------------------------------------------------------
// ListEntries
// ---------------------------------------------------------------------------

func TestListEntries(t *testing.T) {
	entries := &mock.EntryRepository{
		ListAccessibleFn: func(ctx context.Context, userID, pathPrefix string, limit, offset int) ([]*domain.Entry, int, error) {
			return []*domain.Entry{
				{ID: "e1", Title: "Entry 1"},
				{ID: "e2", Title: "Entry 2"},
			}, 2, nil
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, &mock.ObjectStorage{}, &mock.Indexer{}, &mock.Embedder{}, 65536)
	results, total, err := svc.ListEntries(context.Background(), "user-1", "", 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 2, total)
}
