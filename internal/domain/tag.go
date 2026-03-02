package domain

import (
	"context"
	"time"
)

// Tag represents a label that can be attached to entries.
type Tag struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
}

// TagRepository defines persistence operations for tags.
type TagRepository interface {
	Create(ctx context.Context, tag *Tag) error
	GetByID(ctx context.Context, id string) (*Tag, error)
	GetBySlug(ctx context.Context, slug string) (*Tag, error)
	List(ctx context.Context, limit, offset int) ([]*Tag, int, error)
	Delete(ctx context.Context, id string) error
	AttachToEntry(ctx context.Context, entryID, tagID string) error
	DetachFromEntry(ctx context.Context, entryID, tagID string) error
	ListByEntry(ctx context.Context, entryID string) ([]*Tag, error)
}
