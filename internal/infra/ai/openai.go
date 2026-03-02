package ai

import (
	"context"
	"fmt"
	"strings"

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
	}

	return &OpenAIProvider{
		client:         client,
		embeddingModel: openai.EmbeddingModel(cfg.EmbeddingModel),
		chatModel:      cfg.ChatModel,
	}
}

func (p *OpenAIProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if p.client == nil {
		// No-op: return zero vector when API key is not configured.
		return make([]float32, 1536), nil
	}

	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: p.embeddingModel,
	})
	if err != nil {
		return nil, fmt.Errorf("creating embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

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
		return "", nil, nil
	}

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.chatModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf(summarizePrompt, text),
			},
		},
		Temperature: 0.3,
		MaxCompletionTokens: 300,
	})
	if err != nil {
		return "", nil, fmt.Errorf("creating chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", nil, fmt.Errorf("empty chat completion response")
	}

	return parseSummarizeResponse(resp.Choices[0].Message.Content)
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
