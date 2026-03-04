package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/infra/postgres/query"
)

// TagRepository implements domain.TagRepository using PostgreSQL.
type TagRepository struct {
	q *query.Queries
}

// NewTagRepository creates a new TagRepository backed by the given connection pool.
func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{q: query.New(pool)}
}

func (r *TagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	row, err := r.q.InsertTag(ctx, query.InsertTagParams{
		Name: tag.Name,
		Slug: tag.Slug,
	})
	if err != nil {
		return fmt.Errorf("inserting tag: %w", err)
	}

	mapTagRow(&row, tag)
	return nil
}

func (r *TagRepository) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("parsing tag id: %w", err)
	}

	row, err := r.q.GetTagByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("getting tag: %w", err)
	}

	tag := &domain.Tag{}
	mapTagRow(&row, tag)
	return tag, nil
}

func (r *TagRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tag, error) {
	row, err := r.q.GetTagBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("getting tag by slug: %w", err)
	}

	tag := &domain.Tag{}
	mapTagRow(&row, tag)
	return tag, nil
}

func (r *TagRepository) List(ctx context.Context, limit, offset int) ([]*domain.Tag, int, error) {
	count, err := r.q.CountTags(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("counting tags: %w", err)
	}

	rows, err := r.q.ListTags(ctx, query.ListTagsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("listing tags: %w", err)
	}

	tags := make([]*domain.Tag, len(rows))
	for i := range rows {
		tags[i] = &domain.Tag{}
		mapTagRow(&rows[i], tags[i])
	}
	return tags, int(count), nil
}

func (r *TagRepository) Delete(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return fmt.Errorf("parsing tag id: %w", err)
	}

	if err := r.q.DeleteTag(ctx, uid); err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	return nil
}

func (r *TagRepository) AttachToEntry(ctx context.Context, entryID, tagID string) error {
	eID, err := parseUUID(entryID)
	if err != nil {
		return fmt.Errorf("parsing entry id: %w", err)
	}

	tID, err := parseUUID(tagID)
	if err != nil {
		return fmt.Errorf("parsing tag id: %w", err)
	}

	if err := r.q.AttachTagToEntry(ctx, query.AttachTagToEntryParams{
		EntryID: eID,
		TagID:   tID,
	}); err != nil {
		return fmt.Errorf("attaching tag: %w", err)
	}
	return nil
}

func (r *TagRepository) DetachFromEntry(ctx context.Context, entryID, tagID string) error {
	eID, err := parseUUID(entryID)
	if err != nil {
		return fmt.Errorf("parsing entry id: %w", err)
	}

	tID, err := parseUUID(tagID)
	if err != nil {
		return fmt.Errorf("parsing tag id: %w", err)
	}

	if err := r.q.DetachTagFromEntry(ctx, query.DetachTagFromEntryParams{
		EntryID: eID,
		TagID:   tID,
	}); err != nil {
		return fmt.Errorf("detaching tag: %w", err)
	}
	return nil
}

func (r *TagRepository) ListByEntry(ctx context.Context, entryID string) ([]*domain.Tag, error) {
	eID, err := parseUUID(entryID)
	if err != nil {
		return nil, fmt.Errorf("parsing entry id: %w", err)
	}

	rows, err := r.q.ListTagsByEntry(ctx, eID)
	if err != nil {
		return nil, fmt.Errorf("listing tags by entry: %w", err)
	}

	tags := make([]*domain.Tag, len(rows))
	for i := range rows {
		tags[i] = &domain.Tag{}
		mapTagRow(&rows[i], tags[i])
	}
	return tags, nil
}

func (r *TagRepository) ListByEntries(ctx context.Context, entryIDs []string) (map[string][]string, error) {
	uuids := make([]pgtype.UUID, 0, len(entryIDs))
	for _, id := range entryIDs {
		uid, err := parseUUID(id)
		if err != nil {
			return nil, fmt.Errorf("parsing entry id: %w", err)
		}
		uuids = append(uuids, uid)
	}

	rows, err := r.q.ListTagsByEntries(ctx, uuids)
	if err != nil {
		return nil, fmt.Errorf("listing tags by entries: %w", err)
	}

	result := make(map[string][]string, len(entryIDs))
	for _, row := range rows {
		eid := uuidToString(row.EntryID)
		result[eid] = append(result[eid], row.Name)
	}
	return result, nil
}

// mapTagRow maps a sqlc-generated Tag row to a domain Tag.
func mapTagRow(row *query.Tag, tag *domain.Tag) {
	tag.ID = uuidToString(row.ID)
	tag.Name = row.Name
	tag.Slug = row.Slug
	tag.CreatedAt = row.CreatedAt.Time
}
