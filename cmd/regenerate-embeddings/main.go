// cmd/regenerate-embeddings recomputes OpenAI embeddings for all entries and
// saves them to the database. It does NOT index to Meilisearch — run
// cmd/reindex afterwards to push the new embeddings into the search index.
//
// Usage:
//
//	go run ./cmd/regenerate-embeddings [-config config.yaml] [-batch-delay 200ms] [-dry-run]
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/emirakts0/mahzen/internal/config"
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

	entries, err := postgres.NewEntryRepository(pool).ListAll(ctx)
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

		embedding, err := embedder.Embed(ctx, entry.EmbedText())
		if err != nil {
			slog.Error("embed failed", "entry_id", entry.ID, "error", err)
			failed++
			continue
		}

		if *dryRun {
			ok++
			continue
		}

		if err := postgres.NewEntryRepository(pool).UpdateEmbedding(ctx, entry.ID, embedding); err != nil {
			slog.Error("failed to save embedding to db", "entry_id", entry.ID, "error", err)
			failed++
			continue
		}

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
