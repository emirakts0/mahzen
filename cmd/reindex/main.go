// cmd/reindex reads all entries from the database (with existing embeddings)
// and indexes them to Meilisearch. It does NOT regenerate embeddings.
//
// Usage:
//
//	go run ./cmd/reindex [-config config.yaml] [-batch-delay 50ms] [-dry-run]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emirakts0/mahzen/internal/config"
	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/infra/meilisearch"
	"github.com/emirakts0/mahzen/internal/infra/postgres"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	batchDelay := flag.Duration("batch-delay", 50*time.Millisecond, "delay between Meilisearch upserts")
	dryRun := flag.Bool("dry-run", false, "list entries but do not index to Meilisearch")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Connect to Postgres.
	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to postgres: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to postgres")

	// Connect to Meilisearch.
	meilClient, err := meilisearch.NewClient(cfg.Meilisearch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to meilisearch: %v\n", err)
		os.Exit(1)
	}
	slog.Info("connected to meilisearch")

	// Ensure index exists.
	if err := meilisearch.EnsureIndex(ctx, meilClient); err != nil {
		fmt.Fprintf(os.Stderr, "ensure meilisearch index: %v\n", err)
		os.Exit(1)
	}

	indexer := meilisearch.NewIndexer(meilClient)

	// Fetch all entries from Postgres.
	entries, err := listAllEntries(ctx, pool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list entries: %v\n", err)
		os.Exit(1)
	}
	slog.Info("found entries", "count", len(entries))

	ok, failed, noEmbedding := 0, 0, 0
	for i, entry := range entries {
		slog.Info("processing entry",
			"index", i+1,
			"total", len(entries),
			"entry_id", entry.ID,
			"title", entry.Title,
			"has_embedding", entry.Embedding != nil,
		)

		if *dryRun {
			slog.Info("dry-run: skipping meilisearch upsert", "entry_id", entry.ID)
			if entry.Embedding == nil {
				noEmbedding++
			} else {
				ok++
			}
			continue
		}

		// Fetch tags for this entry.
		tags, err := listTagsForEntry(ctx, pool, entry.ID)
		if err != nil {
			slog.Warn("failed to fetch tags", "entry_id", entry.ID, "error", err)
			tags = nil
		}

		if err := indexer.IndexEntry(ctx, entry, tags, entry.Embedding); err != nil {
			slog.Error("meilisearch upsert failed", "entry_id", entry.ID, "error", err)
			failed++
			continue
		}

		if entry.Embedding == nil {
			noEmbedding++
		}
		slog.Info("entry indexed", "entry_id", entry.ID)
		ok++

		if *batchDelay > 0 && i < len(entries)-1 {
			time.Sleep(*batchDelay)
		}
	}

	slog.Info("reindex complete",
		"ok", ok,
		"failed", failed,
		"no_embedding", noEmbedding,
		"total", len(entries),
	)
	if failed > 0 {
		os.Exit(1)
	}
}

func listAllEntries(ctx context.Context, pool *pgxpool.Pool) ([]*domain.Entry, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, user_id, title, content, summary, path, visibility, file_type, file_size, embedding, created_at, updated_at
		FROM entries
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.Entry
	for rows.Next() {
		var (
			id, userID                         pgtype.UUID
			title, content, summary, path, vis string
			fileType                           string
			fileSize                           int64
			embeddingJSON                      *string
			createdAt, updatedAt               pgtype.Timestamptz
		)
		if err := rows.Scan(&id, &userID, &title, &content, &summary, &path, &vis, &fileType, &fileSize, &embeddingJSON, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scanning entry row: %w", err)
		}

		var embedding []float32
		if embeddingJSON != nil && *embeddingJSON != "" {
			if err := json.Unmarshal([]byte(*embeddingJSON), &embedding); err != nil {
				slog.Warn("failed to parse embedding json", "entry_id", uuidStr(id), "error", err)
			}
		}

		entries = append(entries, &domain.Entry{
			ID:         uuidStr(id),
			UserID:     uuidStr(userID),
			Title:      title,
			Content:    content,
			Summary:    summary,
			Path:       path,
			Visibility: domain.ParseVisibility(vis),
			FileType:   fileType,
			FileSize:   fileSize,
			Embedding:  embedding,
			CreatedAt:  createdAt.Time,
			UpdatedAt:  updatedAt.Time,
		})
	}
	return entries, rows.Err()
}

func listTagsForEntry(ctx context.Context, pool *pgxpool.Pool, entryID string) ([]*domain.Tag, error) {
	rows, err := pool.Query(ctx, `
		SELECT t.id, t.name, t.slug, t.created_at
		FROM tags t
		JOIN entry_tags et ON et.tag_id = t.id
		WHERE et.entry_id = $1
		ORDER BY t.name
	`, entryID)
	if err != nil {
		return nil, fmt.Errorf("querying tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		var (
			id         pgtype.UUID
			name, slug string
			createdAt  pgtype.Timestamptz
		)
		if err := rows.Scan(&id, &name, &slug, &createdAt); err != nil {
			return nil, fmt.Errorf("scanning tag row: %w", err)
		}
		tags = append(tags, &domain.Tag{
			ID:        uuidStr(id),
			Name:      name,
			Slug:      slug,
			CreatedAt: createdAt.Time,
		})
	}
	return tags, rows.Err()
}

func uuidStr(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}
