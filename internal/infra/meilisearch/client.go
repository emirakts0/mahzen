package meilisearch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	meili "github.com/meilisearch/meilisearch-go"

	"github.com/emirakts0/mahzen/internal/config"
)

const (
	IndexName    = "entries"
	EmbeddingDim = 1536
)

// NewClient creates and returns a Meilisearch client configured from the application config.
func NewClient(cfg config.MeilisearchConfig) (meili.ServiceManager, error) {
	host := cfg.Host
	if host != "" && !containsScheme(host) {
		host = "http://" + host
	}

	client := meili.New(host, meili.WithAPIKey(cfg.APIKey))

	// Verify connectivity with a health check.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := client.HealthWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("meilisearch health check failed: %w", err)
	}
	if health.Status != "available" {
		return nil, fmt.Errorf("meilisearch health check returned status: %s", health.Status)
	}

	return client, nil
}

// EnsureIndex creates the entries index if it does not already exist and configures
// searchable/filterable attributes and the userProvided embedder for vector search.
func EnsureIndex(ctx context.Context, client meili.ServiceManager) error {
	idx := client.Index(IndexName)

	// Check if index already exists by getting its stats.
	if _, err := idx.GetStatsWithContext(ctx); err == nil {
		slog.Info("meilisearch index already exists", "index", IndexName)
		return configureSettings(ctx, idx)
	}

	// Create the index.
	if _, err := client.CreateIndexWithContext(ctx, &meili.IndexConfig{Uid: IndexName, PrimaryKey: "id"}); err != nil {
		return fmt.Errorf("creating meilisearch index %s: %w", IndexName, err)
	}

	slog.Info("meilisearch index created", "index", IndexName)
	return configureSettings(ctx, idx)
}

func configureSettings(ctx context.Context, idx meili.IndexManager) error {
	settings := &meili.Settings{
		SearchableAttributes: []string{"title", "content", "summary", "tags"},
		DisplayedAttributes: []string{
			"id", "user_id", "title", "summary", "content",
			"path", "visibility", "tags", "created_at", "file_type", "file_size",
		},
		SortableAttributes: []string{"created_at"},
		RankingRules: []string{
			"words", "typo", "proximity", "attribute",
			"sort", "exactness", "created_at:desc",
		},
		StopWords: []string{
			"a", "an", "the", "and", "or", "is", "are", "was", "were",
			"be", "been", "being", "in", "on", "at", "to", "of", "for",
			"it", "its", "this", "that",
		},
		TypoTolerance: &meili.TypoTolerance{
			Enabled:          true,
			DisableOnNumbers: true,
			MinWordSizeForTypos: meili.MinWordSizeForTypos{
				OneTypo:  5,
				TwoTypos: 9,
			},
		},
		Pagination: &meili.Pagination{
			MaxTotalHits: 1000,
		},
		SearchCutoffMs: 1500,
		Embedders: map[string]meili.Embedder{
			"openai": {
				Source:     meili.UserProvidedEmbedderSource,
				Dimensions: EmbeddingDim,
			},
		},
	}

	task, err := idx.UpdateSettingsWithContext(ctx, settings)
	if err != nil {
		return fmt.Errorf("updating meilisearch index settings: %w", err)
	}

	if _, err := idx.WaitForTask(task.TaskUID, 10*time.Second); err != nil {
		return fmt.Errorf("waiting for meilisearch settings update: %w", err)
	}

	// Set filterable attributes separately using the dedicated endpoint
	// to leverage AttributeRule for per-field feature optimization.
	filterableAttrs := &[]interface{}{
		"user_id",
		"visibility",
		"tags",
		"path",
		meili.AttributeRule{
			AttributePatterns: []string{"created_at"},
			Features: meili.AttributeFeatures{
				Filter: meili.FilterFeatures{
					Equality:   false,
					Comparison: true,
				},
			},
		},
	}

	faTask, err := idx.UpdateFilterableAttributesWithContext(ctx, filterableAttrs)
	if err != nil {
		return fmt.Errorf("updating meilisearch filterable attributes: %w", err)
	}

	if _, err := idx.WaitForTask(faTask.TaskUID, 10*time.Second); err != nil {
		return fmt.Errorf("waiting for meilisearch filterable attributes update: %w", err)
	}

	slog.Info("meilisearch index settings configured", "index", IndexName)
	return nil
}

func containsScheme(host string) bool {
	return len(host) > 7 && (host[:7] == "http://" || host[:8] == "https://")
}
