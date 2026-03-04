// cmd/reindex is a one-shot admin tool that regenerates OpenAI embeddings for
// all entries and upserts them into Typesense. Run after the OpenAI API key is
// configured or whenever embeddings need to be refreshed.
//
// Usage:
//
//	go run ./cmd/reindex [-config config.yaml] [-batch 5] [-dry-run]
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emirakts0/mahzen/internal/config"
	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/infra/ai"
	"github.com/emirakts0/mahzen/internal/infra/postgres"
	"github.com/emirakts0/mahzen/internal/infra/typesense"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	batchDelay := flag.Duration("batch-delay", 200*time.Millisecond, "delay between OpenAI calls to avoid rate limits")
	dryRun := flag.Bool("dry-run", false, "generate embeddings but do not write to Typesense")
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

	// Connect to Typesense.
	tsClient, err := typesense.NewClient(cfg.Typesense)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to typesense: %v\n", err)
		os.Exit(1)
	}
	slog.Info("connected to typesense")

	// Build infra components.
	indexer := typesense.NewIndexer(tsClient)
	embedder, _ := ai.NewProvider(cfg.OpenAI)

	// Fetch all entries from Postgres.
	entries, err := listAllEntries(ctx, pool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list entries: %v\n", err)
		os.Exit(1)
	}
	slog.Info("found entries", "count", len(entries))

	ok, failed := 0, 0
	for i, entry := range entries {
		slog.Info("processing entry",
			"index", i+1,
			"total", len(entries),
			"entry_id", entry.ID,
			"title", entry.Title,
		)

		// Fetch tags for this entry.
		tags, err := listTagsForEntry(ctx, pool, entry.ID)
		if err != nil {
			slog.Warn("failed to fetch tags", "entry_id", entry.ID, "error", err)
			tags = nil
		}

		// Build embed text: content first, fall back to title.
		embedText := entry.Content
		if embedText == "" {
			embedText = entry.Title
			if entry.Summary != "" {
				embedText += " " + entry.Summary
			}
		}

		embedding, err := embedder.Embed(ctx, embedText)
		if err != nil {
			slog.Error("embed failed", "entry_id", entry.ID, "error", err)
			failed++
			continue
		}
		slog.Info("embedding generated", "entry_id", entry.ID, "dimensions", len(embedding))

		if *dryRun {
			slog.Info("dry-run: skipping typesense upsert", "entry_id", entry.ID)
			ok++
			continue
		}

		if err := indexer.UpdateEntry(ctx, entry, tags, embedding); err != nil {
			slog.Error("typesense upsert failed", "entry_id", entry.ID, "error", err)
			failed++
			continue
		}

		slog.Info("entry re-indexed", "entry_id", entry.ID)
		ok++

		if *batchDelay > 0 && i < len(entries)-1 {
			time.Sleep(*batchDelay)
		}
	}

	slog.Info("reindex complete", "ok", ok, "failed", failed, "total", len(entries))
	if failed > 0 {
		os.Exit(1)
	}
}

// listAllEntries fetches every entry from Postgres regardless of user or visibility.
func listAllEntries(ctx context.Context, pool *pgxpool.Pool) ([]*domain.Entry, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, user_id, title, content, summary, s3_key, path, visibility, created_at, updated_at
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
			id, userID                                pgtype.UUID
			title, content, summary, s3key, path, vis string
			createdAt, updatedAt                      pgtype.Timestamptz
		)
		if err := rows.Scan(&id, &userID, &title, &content, &summary, &s3key, &path, &vis, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scanning entry row: %w", err)
		}
		entries = append(entries, &domain.Entry{
			ID:         uuidStr(id),
			UserID:     uuidStr(userID),
			Title:      title,
			Content:    content,
			Summary:    summary,
			S3Key:      s3key,
			Path:       path,
			Visibility: domain.ParseVisibility(vis),
			CreatedAt:  createdAt.Time,
			UpdatedAt:  updatedAt.Time,
		})
	}
	return entries, rows.Err()
}

// listTagsForEntry returns the tags attached to a given entry.
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

// uuidStr converts a pgtype.UUID to its string representation.
func uuidStr(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}
