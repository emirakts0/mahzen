// cmd/reindex rebuilds the Meilisearch index from PostgreSQL using the
// embeddings already stored in the database. It does NOT regenerate embeddings.
//
// Usage:
//
//	go run ./cmd/reindex [-config config.yaml] [-batch-delay 50ms] [-dry-run]
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/emirakts0/mahzen/internal/config"
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

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to postgres: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to postgres")

	meilClient, err := meilisearch.NewClient(cfg.Meilisearch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to meilisearch: %v\n", err)
		os.Exit(1)
	}
	if err := meilisearch.EnsureIndex(ctx, meilClient); err != nil {
		fmt.Fprintf(os.Stderr, "ensure meilisearch index: %v\n", err)
		os.Exit(1)
	}
	slog.Info("connected to meilisearch")

	entries, err := postgres.NewEntryRepository(pool).ListAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list entries: %v\n", err)
		os.Exit(1)
	}
	slog.Info("found entries", "count", len(entries))

	tags := postgres.NewTagRepository(pool)
	indexer := meilisearch.NewIndexer(meilClient)

	ok, failed, noEmbedding := 0, 0, 0
	for i, entry := range entries {
		slog.Info("processing entry",
			"index", i+1,
			"total", len(entries),
			"entry_id", entry.ID,
			"title", entry.Title,
			"has_embedding", entry.Embedding != nil,
		)

		if entry.Embedding == nil {
			noEmbedding++
		}
		if *dryRun {
			ok++
			continue
		}

		entryTags, err := tags.ListByEntry(ctx, entry.ID)
		if err != nil {
			slog.Warn("failed to fetch tags", "entry_id", entry.ID, "error", err)
		}

		if err := indexer.IndexEntry(ctx, entry, entryTags, entry.Embedding); err != nil {
			slog.Error("meilisearch upsert failed", "entry_id", entry.ID, "error", err)
			failed++
			continue
		}

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
