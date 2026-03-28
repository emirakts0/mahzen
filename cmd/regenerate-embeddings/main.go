// cmd/regenerate-embeddings regenerates OpenAI embeddings for all entries
// and saves them to the database. It does NOT index to Typesense.
//
// Usage:
//
//	go run ./cmd/regenerate-embeddings [-config config.yaml] [-batch-delay 200ms] [-dry-run]
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
	"github.com/emirakts0/mahzen/internal/infra/ai"
	"github.com/emirakts0/mahzen/internal/infra/postgres"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	batchDelay := flag.Duration("batch-delay", 200*time.Millisecond, "delay between OpenAI calls to avoid rate limits")
	dryRun := flag.Bool("dry-run", false, "generate embeddings but do not write to database")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to postgres: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to postgres")

	embedder, _ := ai.NewProvider(cfg.OpenAI)

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

		embedText := entry.EmbedText()

		embedding, err := embedder.Embed(ctx, embedText)
		if err != nil {
			slog.Error("embed failed", "entry_id", entry.ID, "error", err)
			failed++
			continue
		}
		slog.Info("embedding generated", "entry_id", entry.ID, "dimensions", len(embedding))

		if *dryRun {
			slog.Info("dry-run: skipping db save", "entry_id", entry.ID)
			ok++
			continue
		}

		if err := saveEmbeddingToDB(ctx, pool, entry.ID, embedding); err != nil {
			slog.Error("failed to save embedding to db", "entry_id", entry.ID, "error", err)
			failed++
			continue
		}

		slog.Info("embedding saved to db", "entry_id", entry.ID)
		ok++

		if *batchDelay > 0 && i < len(entries)-1 {
			time.Sleep(*batchDelay)
		}
	}

	slog.Info("regenerate complete", "ok", ok, "failed", failed, "total", len(entries))
	if failed > 0 {
		os.Exit(1)
	}
}

func saveEmbeddingToDB(ctx context.Context, pool *pgxpool.Pool, entryID string, embedding []float32) error {
	data, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("marshaling embedding: %w", err)
	}
	_, err = pool.Exec(ctx, `UPDATE entries SET embedding = $1, updated_at = now() WHERE id = $2`, string(data), entryID)
	return err
}

func listAllEntries(ctx context.Context, pool *pgxpool.Pool) ([]*domain.Entry, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, user_id, title, content, summary, path, visibility, file_type, file_size, created_at, updated_at
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
			createdAt, updatedAt               pgtype.Timestamptz
		)
		if err := rows.Scan(&id, &userID, &title, &content, &summary, &path, &vis, &fileType, &fileSize, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scanning entry row: %w", err)
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
			CreatedAt:  createdAt.Time,
			UpdatedAt:  updatedAt.Time,
		})
	}
	return entries, rows.Err()
}

func uuidStr(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}
