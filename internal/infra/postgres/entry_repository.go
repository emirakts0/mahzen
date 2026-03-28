package postgres

import (
	"context"
	"encoding/json"
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
		Path:       entry.Path,
		Visibility: entry.Visibility.String(),
		FileType:   entry.FileType,
		FileSize:   entry.FileSize,
		Embedding:  embeddingToText(entry.Embedding),
	})
	if err != nil {
		return fmt.Errorf("inserting entry: %w", err)
	}

	entry.ID = uuidToString(row.ID)
	entry.UserID = uuidToString(row.UserID)
	entry.Title = row.Title
	entry.Content = row.Content
	entry.Summary = row.Summary
	entry.Path = row.Path
	entry.Visibility = domain.ParseVisibility(row.Visibility)
	entry.FileType = row.FileType
	entry.FileSize = row.FileSize
	entry.Embedding = textToEmbedding(row.Embedding)
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
		Path:       row.Path,
		Visibility: domain.ParseVisibility(row.Visibility),
		FileType:   row.FileType,
		FileSize:   row.FileSize,
		Embedding:  textToEmbedding(row.Embedding),
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
		Path:       entry.Path,
		Visibility: entry.Visibility.String(),
		FileType:   entry.FileType,
		FileSize:   entry.FileSize,
		Embedding:  embeddingToText(entry.Embedding),
	})
	if err != nil {
		return fmt.Errorf("updating entry: %w", err)
	}

	entry.ID = uuidToString(row.ID)
	entry.UserID = uuidToString(row.UserID)
	entry.Title = row.Title
	entry.Content = row.Content
	entry.Summary = row.Summary
	entry.Path = row.Path
	entry.Visibility = domain.ParseVisibility(row.Visibility)
	entry.FileType = row.FileType
	entry.FileSize = row.FileSize
	entry.Embedding = textToEmbedding(row.Embedding)
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
			Path:       row.Path,
			Visibility: domain.ParseVisibility(row.Visibility),
			FileType:   row.FileType,
			FileSize:   row.FileSize,
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
				Path:       row.Path,
				Visibility: domain.ParseVisibility(row.Visibility),
				FileType:   row.FileType,
				FileSize:   row.FileSize,
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
			Path:       row.Path,
			Visibility: domain.ParseVisibility(row.Visibility),
			FileType:   row.FileType,
			FileSize:   row.FileSize,
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		}
	}
	return entries, int(count), nil
}

func (r *EntryRepository) ListDistinctPaths(ctx context.Context, userID string) ([]string, error) {
	var uid pgtype.UUID
	if userID != "" {
		var err error
		uid, err = parseUUID(userID)
		if err != nil {
			return nil, fmt.Errorf("parsing user id: %w", err)
		}
	}

	paths, err := r.q.ListDistinctPaths(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("listing distinct paths: %w", err)
	}

	return paths, nil
}

func (r *EntryRepository) ListInPath(ctx context.Context, userID, path string, own bool, filter *domain.ListEntriesFilter, limit, offset int) ([]*domain.Entry, int, error) {
	var uid pgtype.UUID
	if userID != "" {
		var err error
		uid, err = parseUUID(userID)
		if err != nil {
			return nil, 0, fmt.Errorf("parsing user id: %w", err)
		}
	}

	filterVisibility, fromDate, toDate, filterTags := extractFilterParams(filter)

	entries, err := r.q.ListEntriesInPath(ctx, query.ListEntriesInPathParams{
		UserID:           uid,
		Path:             path,
		Limit:            int32(limit),
		Offset:           int32(offset),
		Own:              own,
		FilterVisibility: filterVisibility,
		FromDate:         fromDate,
		ToDate:           toDate,
		FilterTags:       filterTags,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("listing entries in path: %w", err)
	}

	count, err := r.q.CountEntriesInPath(ctx, query.CountEntriesInPathParams{
		UserID:           uid,
		Path:             path,
		Own:              own,
		FilterVisibility: filterVisibility,
		FromDate:         fromDate,
		ToDate:           toDate,
		FilterTags:       filterTags,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("counting entries in path: %w", err)
	}

	result := make([]*domain.Entry, len(entries))
	for i, e := range entries {
		result[i] = &domain.Entry{
			ID:         uuidToString(e.ID),
			UserID:     uuidToString(e.UserID),
			Title:      e.Title,
			Content:    e.Content,
			Summary:    e.Summary,
			Path:       e.Path,
			Visibility: domain.ParseVisibility(e.Visibility),
			FileType:   e.FileType,
			FileSize:   e.FileSize,
			CreatedAt:  e.CreatedAt.Time,
			UpdatedAt:  e.UpdatedAt.Time,
		}
	}

	return result, int(count), nil
}

func (r *EntryRepository) ListPathsUnderPrefix(ctx context.Context, userID, prefix string, own bool, filter *domain.ListEntriesFilter) ([]string, error) {
	var uid pgtype.UUID
	if userID != "" {
		var err error
		uid, err = parseUUID(userID)
		if err != nil {
			return nil, fmt.Errorf("parsing user id: %w", err)
		}
	}

	filterVisibility, fromDate, toDate, filterTags := extractFilterParams(filter)

	var paths []string
	var err error

	if prefix == "/" {
		paths, err = r.q.ListAllPaths(ctx, query.ListAllPathsParams{
			UserID:           uid,
			Own:              own,
			FilterVisibility: filterVisibility,
			FromDate:         fromDate,
			ToDate:           toDate,
			FilterTags:       filterTags,
		})
	} else {
		paths, err = r.q.ListPathsUnderPrefix(ctx, query.ListPathsUnderPrefixParams{
			UserID:           uid,
			Prefix:           prefix,
			Own:              own,
			FilterVisibility: filterVisibility,
			FromDate:         fromDate,
			ToDate:           toDate,
			FilterTags:       filterTags,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("listing paths under prefix: %w", err)
	}

	return paths, nil
}

func extractFilterParams(filter *domain.ListEntriesFilter) (string, pgtype.Timestamptz, pgtype.Timestamptz, []string) {
	var filterVisibility string
	var fromDate, toDate pgtype.Timestamptz
	var filterTags []string

	if filter != nil {
		if filter.Visibility != "" {
			filterVisibility = filter.Visibility
		}
		if !filter.FromDate.IsZero() {
			fromDate = pgtype.Timestamptz{Time: filter.FromDate, Valid: true}
		}
		if !filter.ToDate.IsZero() {
			toDate = pgtype.Timestamptz{Time: filter.ToDate, Valid: true}
		}
		if len(filter.Tags) > 0 {
			filterTags = filter.Tags
		}
	}

	return filterVisibility, fromDate, toDate, filterTags
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

// embeddingToText converts a float32 slice to JSON text for storage.
func embeddingToText(emb []float32) pgtype.Text {
	if len(emb) == 0 {
		return pgtype.Text{Valid: false}
	}
	data, err := json.Marshal(emb)
	if err != nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: string(data), Valid: true}
}

// textToEmbedding parses JSON text into a float32 slice.
func textToEmbedding(t pgtype.Text) []float32 {
	if !t.Valid || t.String == "" {
		return nil
	}
	var emb []float32
	if err := json.Unmarshal([]byte(t.String), &emb); err != nil {
		return nil
	}
	return emb
}

// ListAll returns all entries regardless of user or visibility (for admin/reindex operations).
func (r *EntryRepository) ListAll(ctx context.Context) ([]*domain.Entry, error) {
	rows, err := r.q.ListAllEntries(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing all entries: %w", err)
	}

	entries := make([]*domain.Entry, len(rows))
	for i, row := range rows {
		entries[i] = &domain.Entry{
			ID:         uuidToString(row.ID),
			UserID:     uuidToString(row.UserID),
			Title:      row.Title,
			Content:    row.Content,
			Summary:    row.Summary,
			Path:       row.Path,
			Visibility: domain.ParseVisibility(row.Visibility),
			FileType:   row.FileType,
			FileSize:   row.FileSize,
			Embedding:  textToEmbedding(row.Embedding),
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		}
	}
	return entries, nil
}

// UpdateEmbedding updates only the embedding field for an entry.
func (r *EntryRepository) UpdateEmbedding(ctx context.Context, entryID string, embedding []float32) error {
	uid, err := parseUUID(entryID)
	if err != nil {
		return fmt.Errorf("parsing entry id: %w", err)
	}

	if err := r.q.UpdateEntryEmbedding(ctx, query.UpdateEntryEmbeddingParams{
		ID:        uid,
		Embedding: embeddingToText(embedding),
	}); err != nil {
		return fmt.Errorf("updating entry embedding: %w", err)
	}
	return nil
}

// CountAll returns the total number of entries in the database.
func (r *EntryRepository) CountAll(ctx context.Context) (int64, error) {
	return r.q.CountAllEntries(ctx)
}
