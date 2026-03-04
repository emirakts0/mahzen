package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/emirakts0/mahzen/internal/config"
)

// ---------------------------------------------------------------------------
// No-op provider (empty API key)
// ---------------------------------------------------------------------------

func TestNoOpEmbed(t *testing.T) {
	p := newOpenAI(config.OpenAIConfig{})

	embedding, err := p.Embed(context.Background(), "test text")
	require.NoError(t, err)
	assert.Nil(t, embedding, "no-op embedding should return nil (no API key configured)")
}

func TestNoOpSummarize(t *testing.T) {
	p := newOpenAI(config.OpenAIConfig{})

	summary, tags, err := p.Summarize(context.Background(), "test text")
	require.NoError(t, err)
	assert.Empty(t, summary)
	assert.Nil(t, tags)
}

// ---------------------------------------------------------------------------
// parseSummarizeResponse
// ---------------------------------------------------------------------------

func TestParseSummarizeResponse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSummary string
		wantTags    []string
	}{
		{
			name:        "standard format",
			input:       "SUMMARY: This is a summary.\nTAGS: go, testing, grpc",
			wantSummary: "This is a summary.",
			wantTags:    []string{"go", "testing", "grpc"},
		},
		{
			name:        "extra whitespace",
			input:       "  SUMMARY:   Some summary here.  \n  TAGS:   tag1 ,  tag2  ,  tag3  ",
			wantSummary: "Some summary here.",
			wantTags:    []string{"tag1", "tag2", "tag3"},
		},
		{
			name:        "empty response",
			input:       "",
			wantSummary: "",
			wantTags:    nil,
		},
		{
			name:        "only summary",
			input:       "SUMMARY: Just a summary, no tags.",
			wantSummary: "Just a summary, no tags.",
			wantTags:    nil,
		},
		{
			name:        "only tags",
			input:       "TAGS: tag1, tag2",
			wantSummary: "",
			wantTags:    []string{"tag1", "tag2"},
		},
		{
			name:        "trailing commas in tags",
			input:       "SUMMARY: S\nTAGS: a, b, , c, ",
			wantSummary: "S",
			wantTags:    []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, tags, err := parseSummarizeResponse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantSummary, summary)
			assert.Equal(t, tt.wantTags, tags)
		})
	}
}

// ---------------------------------------------------------------------------
// NewProvider integration
// ---------------------------------------------------------------------------

func TestNewProvider_ReturnsEmbedderAndSummarizer(t *testing.T) {
	embedder, summarizer := NewProvider(config.OpenAIConfig{})
	assert.NotNil(t, embedder)
	assert.NotNil(t, summarizer)
}
