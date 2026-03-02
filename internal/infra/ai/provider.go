package ai

import (
	"github.com/emirakts0/mahzen/internal/config"
	"github.com/emirakts0/mahzen/internal/domain"
)

// NewProvider creates an Embedder and Summarizer from the given config.
// Both interfaces are backed by the same OpenAI client.
func NewProvider(cfg config.OpenAIConfig) (domain.Embedder, domain.Summarizer) {
	p := newOpenAI(cfg)
	return p, p
}
