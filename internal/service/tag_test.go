package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service/mock"
)

// ---------------------------------------------------------------------------
// CreateTag
// ---------------------------------------------------------------------------

func TestCreateTag(t *testing.T) {
	var createdTag *domain.Tag

	tags := &mock.TagRepository{
		CreateFn: func(ctx context.Context, tag *domain.Tag) error {
			tag.ID = "tag-1"
			createdTag = tag
			return nil
		},
	}

	svc := NewTagService(tags)
	tag, err := svc.CreateTag(context.Background(), "My Great Tag!")
	require.NoError(t, err)
	assert.Equal(t, "tag-1", tag.ID)
	assert.Equal(t, "My Great Tag!", tag.Name)
	assert.Equal(t, "my-great-tag", createdTag.Slug)
}

func TestCreateTag_RepoError(t *testing.T) {
	tags := &mock.TagRepository{
		CreateFn: func(ctx context.Context, tag *domain.Tag) error {
			return errors.New("duplicate slug")
		},
	}

	svc := NewTagService(tags)
	_, err := svc.CreateTag(context.Background(), "existing tag")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating tag")
}

// ---------------------------------------------------------------------------
// GetTag
// ---------------------------------------------------------------------------

func TestGetTag(t *testing.T) {
	tags := &mock.TagRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Tag, error) {
			return &domain.Tag{ID: id, Name: "golang", Slug: "golang"}, nil
		},
	}

	svc := NewTagService(tags)
	tag, err := svc.GetTag(context.Background(), "tag-1")
	require.NoError(t, err)
	assert.Equal(t, "golang", tag.Name)
}

func TestGetTag_NotFound(t *testing.T) {
	tags := &mock.TagRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Tag, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewTagService(tags)
	_, err := svc.GetTag(context.Background(), "nonexistent")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// ListTags
// ---------------------------------------------------------------------------

func TestListTags(t *testing.T) {
	tags := &mock.TagRepository{
		ListFn: func(ctx context.Context, limit, offset int) ([]*domain.Tag, int, error) {
			return []*domain.Tag{
				{ID: "t1", Name: "go", Slug: "go"},
				{ID: "t2", Name: "rust", Slug: "rust"},
			}, 5, nil
		},
	}

	svc := NewTagService(tags)
	results, total, err := svc.ListTags(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 5, total)
}

// ---------------------------------------------------------------------------
// DeleteTag
// ---------------------------------------------------------------------------

func TestDeleteTag(t *testing.T) {
	var deletedID string
	tags := &mock.TagRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			deletedID = id
			return nil
		},
	}

	svc := NewTagService(tags)
	err := svc.DeleteTag(context.Background(), "tag-1")
	require.NoError(t, err)
	assert.Equal(t, "tag-1", deletedID)
}

// ---------------------------------------------------------------------------
// AttachTag / DetachTag
// ---------------------------------------------------------------------------

func TestAttachTag(t *testing.T) {
	var attachedEntryID, attachedTagID string
	tags := &mock.TagRepository{
		AttachToEntryFn: func(ctx context.Context, entryID, tagID string) error {
			attachedEntryID = entryID
			attachedTagID = tagID
			return nil
		},
	}

	svc := NewTagService(tags)
	err := svc.AttachTag(context.Background(), "entry-1", "tag-1")
	require.NoError(t, err)
	assert.Equal(t, "entry-1", attachedEntryID)
	assert.Equal(t, "tag-1", attachedTagID)
}

func TestDetachTag(t *testing.T) {
	tags := &mock.TagRepository{
		DetachFromEntryFn: func(ctx context.Context, entryID, tagID string) error {
			return nil
		},
	}

	svc := NewTagService(tags)
	err := svc.DetachTag(context.Background(), "entry-1", "tag-1")
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// ListTagsByEntry
// ---------------------------------------------------------------------------

func TestListTagsByEntry(t *testing.T) {
	tags := &mock.TagRepository{
		ListByEntryFn: func(ctx context.Context, entryID string) ([]*domain.Tag, error) {
			return []*domain.Tag{
				{ID: "t1", Name: "go"},
				{ID: "t2", Name: "testing"},
			}, nil
		},
	}

	svc := NewTagService(tags)
	results, err := svc.ListTagsByEntry(context.Background(), "entry-1")
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

// ---------------------------------------------------------------------------
// slugify (internal function, white-box tests)
// ---------------------------------------------------------------------------

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lowercase", "hello", "hello"},
		{"mixed case", "Hello World", "hello-world"},
		{"special characters", "Go & Rust!", "go-rust"},
		{"leading/trailing spaces", "  spaced  ", "spaced"},
		{"multiple special chars", "a---b___c", "a-b-c"},
		{"unicode", "cafe 日本語", "cafe"},
		{"empty string", "", ""},
		{"only special chars", "!@#$%", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, slugify(tt.input))
		})
	}
}
