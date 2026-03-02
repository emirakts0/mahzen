package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/emirakts0/mahzen/internal/domain"
)

// TagService contains business logic for managing tags.
type TagService struct {
	tags domain.TagRepository
}

// NewTagService creates a new TagService.
func NewTagService(tags domain.TagRepository) *TagService {
	return &TagService{tags: tags}
}

// CreateTag creates a new tag, automatically generating a slug from the name.
func (s *TagService) CreateTag(ctx context.Context, name string) (*domain.Tag, error) {
	tag := &domain.Tag{
		Name: name,
		Slug: slugify(name),
	}

	if err := s.tags.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("creating tag: %w", err)
	}

	return tag, nil
}

// GetTag retrieves a tag by ID.
func (s *TagService) GetTag(ctx context.Context, id string) (*domain.Tag, error) {
	return s.tags.GetByID(ctx, id)
}

// ListTags lists tags with pagination.
func (s *TagService) ListTags(ctx context.Context, limit, offset int) ([]*domain.Tag, int, error) {
	return s.tags.List(ctx, limit, offset)
}

// DeleteTag removes a tag.
func (s *TagService) DeleteTag(ctx context.Context, id string) error {
	return s.tags.Delete(ctx, id)
}

// AttachTag attaches a tag to an entry.
func (s *TagService) AttachTag(ctx context.Context, entryID, tagID string) error {
	return s.tags.AttachToEntry(ctx, entryID, tagID)
}

// DetachTag detaches a tag from an entry.
func (s *TagService) DetachTag(ctx context.Context, entryID, tagID string) error {
	return s.tags.DetachFromEntry(ctx, entryID, tagID)
}

// ListTagsByEntry returns all tags attached to a given entry.
func (s *TagService) ListTagsByEntry(ctx context.Context, entryID string) ([]*domain.Tag, error) {
	return s.tags.ListByEntry(ctx, entryID)
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// slugify converts a string to a URL-friendly slug.
func slugify(s string) string {
	slug := strings.ToLower(strings.TrimSpace(s))
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
