package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/emirakts0/mahzen/internal/config"
)

// OpenAIProvider implements both domain.Embedder and domain.Summarizer using OpenAI.
type OpenAIProvider struct {
	client         *openai.Client
	embeddingModel openai.EmbeddingModel
	chatModel      string
}

// newOpenAI creates a new OpenAI provider. If the API key is empty,
// it returns a no-op provider that returns zero embeddings and empty summaries.
func newOpenAI(cfg config.OpenAIConfig) *OpenAIProvider {
	var client *openai.Client
	if cfg.APIKey != "" {
		client = openai.NewClient(cfg.APIKey)
		slog.Info("openai provider initialized",
			"embedding_model", cfg.EmbeddingModel,
			"chat_model", cfg.ChatModel,
		)
	} else {
		slog.Warn("openai provider running in no-op mode (no API key configured)")
	}

	return &OpenAIProvider{
		client:         client,
		embeddingModel: openai.EmbeddingModel(cfg.EmbeddingModel),
		chatModel:      cfg.ChatModel,
	}
}

func (p *OpenAIProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if p.client == nil {
		slog.Debug("openai embed skipped (no-op mode)", "text_length", len(text))
		return make([]float32, 1536), nil
	}

	slog.Info("openai embed request",
		"text_length", len(text),
		"model", string(p.embeddingModel),
	)

	start := time.Now()
	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: p.embeddingModel,
	})
	duration := time.Since(start)

	if err != nil {
		slog.Error("openai embed failed",
			"duration", duration,
			"error", err,
		)
		return nil, fmt.Errorf("creating embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		slog.Error("openai embed returned empty data", "duration", duration)
		return nil, fmt.Errorf("empty embedding response")
	}

	slog.Info("openai embed response",
		"duration", duration,
		"dimensions", len(resp.Data[0].Embedding),
		"usage_tokens", resp.Usage.TotalTokens,
	)

	return resp.Data[0].Embedding, nil
}

const summarizePrompt = `You are a knowledge management assistant. Given the following content, provide:
1. A concise summary (2-3 sentences max).
2. A list of 3-5 relevant tags (single words or short phrases, lowercase).

Respond in exactly this format:
SUMMARY: <your summary>
TAGS: tag1, tag2, tag3

Content:
%s`

func (p *OpenAIProvider) Summarize(ctx context.Context, text string) (string, []string, error) {
	if p.client == nil {
		slog.Debug("openai summarize skipped (no-op mode)", "text_length", len(text))
		return "", nil, nil
	}

	slog.Info("openai summarize request",
		"text_length", len(text),
		"model", p.chatModel,
	)

	start := time.Now()
	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.chatModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf(summarizePrompt, text),
			},
		},
		Temperature:         0.3,
		MaxCompletionTokens: 300,
	})
	duration := time.Since(start)

	if err != nil {
		slog.Error("openai summarize failed",
			"duration", duration,
			"error", err,
		)
		return "", nil, fmt.Errorf("creating chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		slog.Error("openai summarize returned empty choices", "duration", duration)
		return "", nil, fmt.Errorf("empty chat completion response")
	}

	rawResponse := resp.Choices[0].Message.Content
	summary, tags, parseErr := parseSummarizeResponse(rawResponse)

	slog.Info("openai summarize response",
		"duration", duration,
		"summary_length", len(summary),
		"tag_count", len(tags),
		"tags", tags,
		"usage_prompt_tokens", resp.Usage.PromptTokens,
		"usage_completion_tokens", resp.Usage.CompletionTokens,
	)

	return summary, tags, parseErr
}

// parseSummarizeResponse extracts summary and tags from the structured LLM response.
func parseSummarizeResponse(response string) (string, []string, error) {
	var summary string
	var tags []string

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SUMMARY:") {
			summary = strings.TrimSpace(strings.TrimPrefix(line, "SUMMARY:"))
		}
		if strings.HasPrefix(line, "TAGS:") {
			tagStr := strings.TrimSpace(strings.TrimPrefix(line, "TAGS:"))
			for _, t := range strings.Split(tagStr, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tags = append(tags, t)
				}
			}
		}
	}

	return summary, tags, nil
}
