package typesense

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/typesense/typesense-go/v3/typesense"
	"github.com/typesense/typesense-go/v3/typesense/api"
	"github.com/typesense/typesense-go/v3/typesense/api/pointer"

	"github.com/emirakts0/mahzen/internal/config"
)

const (
	CollectionName   = "entries"
	EmbeddingDim     = 1536
)

// NewClient creates and returns a Typesense client configured from the application config.
func NewClient(cfg config.TypesenseConfig) (*typesense.Client, error) {
	client := typesense.NewClient(
		typesense.WithServer(cfg.URL()),
		typesense.WithAPIKey(cfg.APIKey),
		typesense.WithConnectionTimeout(cfg.ConnectionTimeout),
	)

	// Verify connectivity with a health check.
	ok, err := client.Health(context.Background(), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("typesense health check failed: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("typesense health check returned unhealthy")
	}

	return client, nil
}

// EnsureCollections creates the entries collection if it does not already exist.
func EnsureCollections(ctx context.Context, client *typesense.Client) error {
	_, err := client.Collection(CollectionName).Retrieve(ctx)
	if err == nil {
		slog.Info("typesense collection already exists", "collection", CollectionName)
		return nil
	}

	schema := &api.CollectionSchema{
		Name: CollectionName,
		Fields: []api.Field{
			{Name: "entry_id", Type: "string", Facet: pointer.False()},
			{Name: "user_id", Type: "string", Facet: pointer.True()},
			{Name: "title", Type: "string", Facet: pointer.False()},
			{Name: "content", Type: "string", Facet: pointer.False()},
			{Name: "summary", Type: "string", Facet: pointer.False(), Optional: pointer.True()},
			{Name: "path", Type: "string", Facet: pointer.True()},
			{Name: "visibility", Type: "string", Facet: pointer.True()},
			{Name: "tags", Type: "string[]", Facet: pointer.True(), Optional: pointer.True()},
			{Name: "created_at", Type: "int64", Facet: pointer.False()},
			{
				Name:     "embedding",
				Type:     fmt.Sprintf("float[]"),
				NumDim:   pointer.Int(EmbeddingDim),
				Facet:    pointer.False(),
				Optional: pointer.True(),
			},
		},
		DefaultSortingField: pointer.String("created_at"),
	}

	if _, err := client.Collections().Create(ctx, schema); err != nil {
		return fmt.Errorf("creating collection %s: %w", CollectionName, err)
	}

	slog.Info("typesense collection created", "collection", CollectionName)
	return nil
}
