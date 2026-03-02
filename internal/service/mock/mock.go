// Package mock provides hand-rolled test doubles for domain interfaces.
package mock

import (
	"context"
	"io"
	"time"

	"github.com/emirakts0/mahzen/internal/domain"
)

// ---------------------------------------------------------------------------
// EntryRepository
// ---------------------------------------------------------------------------

// EntryRepository is a test double for domain.EntryRepository.
type EntryRepository struct {
	CreateFn         func(ctx context.Context, entry *domain.Entry) error
	GetByIDFn        func(ctx context.Context, id string) (*domain.Entry, error)
	UpdateFn         func(ctx context.Context, entry *domain.Entry) error
	DeleteFn         func(ctx context.Context, id string) error
	ListByUserFn     func(ctx context.Context, userID string, limit, offset int) ([]*domain.Entry, int, error)
	ListAccessibleFn func(ctx context.Context, userID, pathPrefix string, limit, offset int) ([]*domain.Entry, int, error)
}

func (m *EntryRepository) Create(ctx context.Context, entry *domain.Entry) error {
	return m.CreateFn(ctx, entry)
}

func (m *EntryRepository) GetByID(ctx context.Context, id string) (*domain.Entry, error) {
	return m.GetByIDFn(ctx, id)
}

func (m *EntryRepository) Update(ctx context.Context, entry *domain.Entry) error {
	return m.UpdateFn(ctx, entry)
}

func (m *EntryRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *EntryRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Entry, int, error) {
	return m.ListByUserFn(ctx, userID, limit, offset)
}

func (m *EntryRepository) ListAccessible(ctx context.Context, userID, pathPrefix string, limit, offset int) ([]*domain.Entry, int, error) {
	return m.ListAccessibleFn(ctx, userID, pathPrefix, limit, offset)
}

// ---------------------------------------------------------------------------
// TagRepository
// ---------------------------------------------------------------------------

// TagRepository is a test double for domain.TagRepository.
type TagRepository struct {
	CreateFn          func(ctx context.Context, tag *domain.Tag) error
	GetByIDFn         func(ctx context.Context, id string) (*domain.Tag, error)
	GetBySlugFn       func(ctx context.Context, slug string) (*domain.Tag, error)
	ListFn            func(ctx context.Context, limit, offset int) ([]*domain.Tag, int, error)
	DeleteFn          func(ctx context.Context, id string) error
	AttachToEntryFn   func(ctx context.Context, entryID, tagID string) error
	DetachFromEntryFn func(ctx context.Context, entryID, tagID string) error
	ListByEntryFn     func(ctx context.Context, entryID string) ([]*domain.Tag, error)
}

func (m *TagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	return m.CreateFn(ctx, tag)
}

func (m *TagRepository) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
	return m.GetByIDFn(ctx, id)
}

func (m *TagRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tag, error) {
	return m.GetBySlugFn(ctx, slug)
}

func (m *TagRepository) List(ctx context.Context, limit, offset int) ([]*domain.Tag, int, error) {
	return m.ListFn(ctx, limit, offset)
}

func (m *TagRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *TagRepository) AttachToEntry(ctx context.Context, entryID, tagID string) error {
	return m.AttachToEntryFn(ctx, entryID, tagID)
}

func (m *TagRepository) DetachFromEntry(ctx context.Context, entryID, tagID string) error {
	return m.DetachFromEntryFn(ctx, entryID, tagID)
}

func (m *TagRepository) ListByEntry(ctx context.Context, entryID string) ([]*domain.Tag, error) {
	return m.ListByEntryFn(ctx, entryID)
}

// ---------------------------------------------------------------------------
// ObjectStorage
// ---------------------------------------------------------------------------

// ObjectStorage is a test double for domain.ObjectStorage.
type ObjectStorage struct {
	UploadFn          func(ctx context.Context, key string, reader io.Reader, contentType string, size int64) error
	DownloadFn        func(ctx context.Context, key string) (io.ReadCloser, error)
	DeleteFn          func(ctx context.Context, key string) error
	GetPresignedURLFn func(ctx context.Context, key string, expiry time.Duration) (string, error)
}

func (m *ObjectStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) error {
	return m.UploadFn(ctx, key, reader, contentType, size)
}

func (m *ObjectStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return m.DownloadFn(ctx, key)
}

func (m *ObjectStorage) Delete(ctx context.Context, key string) error {
	return m.DeleteFn(ctx, key)
}

func (m *ObjectStorage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return m.GetPresignedURLFn(ctx, key, expiry)
}

// ---------------------------------------------------------------------------
// Indexer
// ---------------------------------------------------------------------------

// Indexer is a test double for domain.Indexer.
type Indexer struct {
	IndexEntryFn  func(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error
	UpdateEntryFn func(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error
	DeleteEntryFn func(ctx context.Context, id string) error
}

func (m *Indexer) IndexEntry(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
	return m.IndexEntryFn(ctx, entry, tags, embedding)
}

func (m *Indexer) UpdateEntry(ctx context.Context, entry *domain.Entry, tags []*domain.Tag, embedding []float32) error {
	return m.UpdateEntryFn(ctx, entry, tags, embedding)
}

func (m *Indexer) DeleteEntry(ctx context.Context, id string) error {
	return m.DeleteEntryFn(ctx, id)
}

// ---------------------------------------------------------------------------
// Searcher
// ---------------------------------------------------------------------------

// Searcher is a test double for domain.Searcher.
type Searcher struct {
	KeywordSearchFn  func(ctx context.Context, query, userID string, limit, offset int) ([]*domain.SearchResult, int, error)
	SemanticSearchFn func(ctx context.Context, embedding []float32, userID string, limit, offset int) ([]*domain.SearchResult, int, error)
}

func (m *Searcher) KeywordSearch(ctx context.Context, query, userID string, limit, offset int) ([]*domain.SearchResult, int, error) {
	return m.KeywordSearchFn(ctx, query, userID, limit, offset)
}

func (m *Searcher) SemanticSearch(ctx context.Context, embedding []float32, userID string, limit, offset int) ([]*domain.SearchResult, int, error) {
	return m.SemanticSearchFn(ctx, embedding, userID, limit, offset)
}

// ---------------------------------------------------------------------------
// Embedder
// ---------------------------------------------------------------------------

// Embedder is a test double for domain.Embedder.
type Embedder struct {
	EmbedFn func(ctx context.Context, text string) ([]float32, error)
}

func (m *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return m.EmbedFn(ctx, text)
}

// ---------------------------------------------------------------------------
// Summarizer
// ---------------------------------------------------------------------------

// Summarizer is a test double for domain.Summarizer.
type Summarizer struct {
	SummarizeFn func(ctx context.Context, text string) (string, []string, error)
}

func (m *Summarizer) Summarize(ctx context.Context, text string) (string, []string, error) {
	return m.SummarizeFn(ctx, text)
}
