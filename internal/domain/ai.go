package domain

import "context"

// Embedder generates vector embeddings from text.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// Summarizer generates summaries and suggests tags from text content.
type Summarizer interface {
	Summarize(ctx context.Context, text string) (summary string, tags []string, err error)
}
