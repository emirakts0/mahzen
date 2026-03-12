package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service/mock"
)

// newTestEntryService constructs an EntryService wired to mocks.
func newTestEntryService(
	entries *mock.EntryRepository,
	tags *mock.TagRepository,
	indexer *mock.Indexer,
	embedder *mock.Embedder,
) *EntryService {
	return NewEntryService(entries, tags, indexer, embedder)
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
		UpdateEmbeddingFn: func(ctx context.Context, entryID string, embedding []float32) error {
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

	svc := newTestEntryService(entries, tags, indexer, embedder)

	entry, err := svc.CreateEntry(context.Background(), "user-1", "Test Title", "short content", "/notes", "", domain.VisibilityPublic, []string{"tag-1"})
	require.NoError(t, err)
	assert.Equal(t, "entry-1", entry.ID)
	assert.Equal(t, "Test Title", entry.Title)
	assert.Equal(t, "short content", entry.Content)
	assert.Equal(t, "/notes", entry.Path)
	assert.Equal(t, domain.VisibilityPublic, entry.Visibility)
}

func TestCreateEntry_RepoError(t *testing.T) {
	entries := &mock.EntryRepository{
		CreateFn: func(ctx context.Context, entry *domain.Entry) error {
			return errors.New("db connection lost")
		},
	}
	tags := &mock.TagRepository{}
	indexer := &mock.Indexer{}
	embedder := &mock.Embedder{}

	svc := newTestEntryService(entries, tags, indexer, embedder)
	_, err := svc.CreateEntry(context.Background(), "user-1", "Title", "content", "/", "", domain.VisibilityPublic, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating entry")
}

func TestCreateEntry_TagAttachError_NonFatal(t *testing.T) {
	entries := &mock.EntryRepository{
		CreateFn: func(ctx context.Context, entry *domain.Entry) error {
			entry.ID = "entry-3"
			return nil
		},
		UpdateEmbeddingFn: func(ctx context.Context, entryID string, embedding []float32) error {
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

	svc := newTestEntryService(entries, tags, indexer, embedder)
	// Tag attach fails but CreateEntry should still succeed.
	entry, err := svc.CreateEntry(context.Background(), "user-1", "Title", "content", "/", "", domain.VisibilityPublic, []string{"bad-tag"})
	require.NoError(t, err)
	assert.Equal(t, "entry-3", entry.ID)
}

// ---------------------------------------------------------------------------
// GetEntry
// ---------------------------------------------------------------------------

func TestGetEntry(t *testing.T) {
	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{
				ID:      id,
				Title:   "My Entry",
				Content: "inline content",
			}, nil
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, &mock.Indexer{}, &mock.Embedder{})

	entry, err := svc.GetEntry(context.Background(), "entry-1")
	require.NoError(t, err)
	assert.Equal(t, "inline content", entry.Content)
}

func TestGetEntry_NotFound(t *testing.T) {
	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, &mock.Indexer{}, &mock.Embedder{})
	_, err := svc.GetEntry(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting entry")
}

// ---------------------------------------------------------------------------
// UpdateEntry
// ---------------------------------------------------------------------------

func TestUpdateEntry(t *testing.T) {
	var updatedEntry *domain.Entry

	entries := &mock.EntryRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Entry, error) {
			return &domain.Entry{ID: id, UserID: "user-1", Content: "old content"}, nil
		},
		UpdateFn: func(ctx context.Context, entry *domain.Entry) error {
			updatedEntry = entry
			return nil
		},
		UpdateEmbeddingFn: func(ctx context.Context, entryID string, embedding []float32) error {
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

	svc := newTestEntryService(entries, tags, indexer, embedder)

	entry, err := svc.UpdateEntry(context.Background(), "entry-1", "Updated Title", "new content", "/projects", "", domain.VisibilityPrivate, nil)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", entry.Title)
	assert.Equal(t, "new content", updatedEntry.Content)
	assert.Equal(t, domain.VisibilityPrivate, entry.Visibility)
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
		UpdateEmbeddingFn: func(ctx context.Context, entryID string, embedding []float32) error {
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

	svc := newTestEntryService(entries, tags, indexer, embedder)
	_, err := svc.UpdateEntry(context.Background(), "entry-1", "Title", "content", "", "", domain.VisibilityPublic, []string{"new-tag-1", "new-tag-2"})
	require.NoError(t, err)
	assert.Equal(t, []string{"old-tag-1"}, detachedTags, "old tags should be detached")
	assert.Equal(t, []string{"new-tag-1", "new-tag-2"}, attachedTags, "new tags should be attached")
}

// ---------------------------------------------------------------------------
// DeleteEntry
// ---------------------------------------------------------------------------

func TestDeleteEntry(t *testing.T) {
	var indexDeleted bool

	entries := &mock.EntryRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	indexer := &mock.Indexer{
		DeleteEntryFn: func(ctx context.Context, id string) error {
			indexDeleted = true
			return nil
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, indexer, &mock.Embedder{})
	err := svc.DeleteEntry(context.Background(), "entry-1")
	require.NoError(t, err)
	assert.True(t, indexDeleted, "search index entry should be deleted")
}

func TestDeleteEntry_NotFound(t *testing.T) {
	entries := &mock.EntryRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return errors.New("not found")
		},
	}

	svc := newTestEntryService(entries, &mock.TagRepository{}, &mock.Indexer{}, &mock.Embedder{})
	err := svc.DeleteEntry(context.Background(), "nonexistent")
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

	svc := newTestEntryService(entries, &mock.TagRepository{}, &mock.Indexer{}, &mock.Embedder{})
	results, total, err := svc.ListEntries(context.Background(), "user-1", "", 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 2, total)
}
