package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/infra/postgres/query"
)

// EntryRepository implements domain.EntryRepository using PostgreSQL.
type EntryRepository struct {
	q *query.Queries
}

// NewEntryRepository creates a new EntryRepository backed by the given connection pool.
func NewEntryRepository(pool *pgxpool.Pool) *EntryRepository {
	return &EntryRepository{q: query.New(pool)}
}

func (r *EntryRepository) Create(ctx context.Context, entry *domain.Entry) error {
	userID, err := parseUUID(entry.UserID)
	if err != nil {
		return fmt.Errorf("parsing user id: %w", err)
	}

	row, err := r.q.InsertEntry(ctx, query.InsertEntryParams{
		UserID:     userID,
		Title:      entry.Title,
		Content:    entry.Content,
		Summary:    entry.Summary,
		S3Key:      entry.S3Key,
		Path:       entry.Path,
		Visibility: entry.Visibility.String(),
	})
	if err != nil {
		return fmt.Errorf("inserting entry: %w", err)
	}

	entry.ID = uuidToString(row.ID)
	entry.UserID = uuidToString(row.UserID)
	entry.Title = row.Title
	entry.Content = row.Content
	entry.Summary = row.Summary
	entry.S3Key = row.S3Key
	entry.Path = row.Path
	entry.Visibility = domain.ParseVisibility(row.Visibility)
	entry.CreatedAt = row.CreatedAt.Time
	entry.UpdatedAt = row.UpdatedAt.Time
	return nil
}

func (r *EntryRepository) GetByID(ctx context.Context, id string) (*domain.Entry, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("parsing entry id: %w", err)
	}

	row, err := r.q.GetEntryByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("getting entry: %w", err)
	}

	return &domain.Entry{
		ID:         uuidToString(row.ID),
		UserID:     uuidToString(row.UserID),
		Title:      row.Title,
		Content:    row.Content,
		Summary:    row.Summary,
		S3Key:      row.S3Key,
		Path:       row.Path,
		Visibility: domain.ParseVisibility(row.Visibility),
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}, nil
}

func (r *EntryRepository) Update(ctx context.Context, entry *domain.Entry) error {
	uid, err := parseUUID(entry.ID)
	if err != nil {
		return fmt.Errorf("parsing entry id: %w", err)
	}

	row, err := r.q.UpdateEntry(ctx, query.UpdateEntryParams{
		ID:         uid,
		Title:      entry.Title,
		Content:    entry.Content,
		Summary:    entry.Summary,
		S3Key:      entry.S3Key,
		Path:       entry.Path,
		Visibility: entry.Visibility.String(),
	})
	if err != nil {
		return fmt.Errorf("updating entry: %w", err)
	}

	entry.ID = uuidToString(row.ID)
	entry.UserID = uuidToString(row.UserID)
	entry.Title = row.Title
	entry.Content = row.Content
	entry.Summary = row.Summary
	entry.S3Key = row.S3Key
	entry.Path = row.Path
	entry.Visibility = domain.ParseVisibility(row.Visibility)
	entry.CreatedAt = row.CreatedAt.Time
	entry.UpdatedAt = row.UpdatedAt.Time
	return nil
}

func (r *EntryRepository) Delete(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return fmt.Errorf("parsing entry id: %w", err)
	}

	if err := r.q.DeleteEntry(ctx, uid); err != nil {
		return fmt.Errorf("deleting entry: %w", err)
	}
	return nil
}

func (r *EntryRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Entry, int, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("parsing user id: %w", err)
	}

	count, err := r.q.CountEntriesByUser(ctx, uid)
	if err != nil {
		return nil, 0, fmt.Errorf("counting entries: %w", err)
	}

	rows, err := r.q.ListEntriesByUser(ctx, query.ListEntriesByUserParams{
		UserID: uid,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("listing entries: %w", err)
	}

	entries := make([]*domain.Entry, len(rows))
	for i, row := range rows {
		entries[i] = &domain.Entry{
			ID:         uuidToString(row.ID),
			UserID:     uuidToString(row.UserID),
			Title:      row.Title,
			Content:    row.Content,
			Summary:    row.Summary,
			S3Key:      row.S3Key,
			Path:       row.Path,
			Visibility: domain.ParseVisibility(row.Visibility),
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		}
	}
	return entries, int(count), nil
}

func (r *EntryRepository) ListAccessible(ctx context.Context, userID, pathPrefix string, limit, offset int) ([]*domain.Entry, int, error) {
	// When userID is empty (unauthenticated request), use a nil UUID.
	// PostgreSQL evaluates "user_id = NULL" as NULL (never true),
	// so only public entries are returned.
	var uid pgtype.UUID
	if userID != "" {
		var err error
		uid, err = parseUUID(userID)
		if err != nil {
			return nil, 0, fmt.Errorf("parsing user id: %w", err)
		}
	}

	// When a path prefix is provided, use the path-filtered queries.
	if pathPrefix != "" {
		count, err := r.q.CountAccessibleEntriesByPath(ctx, query.CountAccessibleEntriesByPathParams{
			UserID: uid,
			Path:   pathPrefix,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("counting accessible entries by path: %w", err)
		}

		rows, err := r.q.ListAccessibleEntriesByPath(ctx, query.ListAccessibleEntriesByPathParams{
			UserID: uid,
			Path:   pathPrefix,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return nil, 0, fmt.Errorf("listing accessible entries by path: %w", err)
		}

		entries := make([]*domain.Entry, len(rows))
		for i, row := range rows {
			entries[i] = &domain.Entry{
				ID:         uuidToString(row.ID),
				UserID:     uuidToString(row.UserID),
				Title:      row.Title,
				Content:    row.Content,
				Summary:    row.Summary,
				S3Key:      row.S3Key,
				Path:       row.Path,
				Visibility: domain.ParseVisibility(row.Visibility),
				CreatedAt:  row.CreatedAt.Time,
				UpdatedAt:  row.UpdatedAt.Time,
			}
		}
		return entries, int(count), nil
	}

	// No path filter — list all accessible entries.
	count, err := r.q.CountAccessibleEntries(ctx, uid)
	if err != nil {
		return nil, 0, fmt.Errorf("counting accessible entries: %w", err)
	}

	rows, err := r.q.ListAccessibleEntries(ctx, query.ListAccessibleEntriesParams{
		UserID: uid,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("listing accessible entries: %w", err)
	}

	entries := make([]*domain.Entry, len(rows))
	for i, row := range rows {
		entries[i] = &domain.Entry{
			ID:         uuidToString(row.ID),
			UserID:     uuidToString(row.UserID),
			Title:      row.Title,
			Content:    row.Content,
			Summary:    row.Summary,
			S3Key:      row.S3Key,
			Path:       row.Path,
			Visibility: domain.ParseVisibility(row.Visibility),
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		}
	}
	return entries, int(count), nil
}

// parseUUID converts a string UUID to pgtype.UUID.
func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return u, fmt.Errorf("invalid uuid %q: %w", s, err)
	}
	return u, nil
}

// uuidToString converts a pgtype.UUID to its string representation.
func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	b := u.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
