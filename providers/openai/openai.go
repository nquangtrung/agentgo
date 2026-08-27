package openai

import (
	"context"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

type OpenAIProvider struct {
	providers.BaseAgentProvider
	response openAIResponsesService
}

func (p OpenAIProvider) GenerateText(ctx context.Context, params providers.AgentProviderPromptMessageParams) (models.LanguageModelOutput, error) {
	resp, err := p.response.New(ctx, responses.ResponseNewParams{
		Model: p.BaseAgentProvider.Context().ModelName,
		Input: convertInputFromParams(params),
	})
	if err != nil {
		return models.LanguageModelOutput{}, err
	}

	return models.LanguageModelOutput{
		Text: resp.OutputText(),
		Usage: models.LanguageModelUsage{
			OutputTokens:    int64(resp.Usage.OutputTokens),
			InputTokens:     int64(resp.Usage.InputTokens),
			CachedTokens:    int64(resp.Usage.InputTokensDetails.CachedTokens) + int64(resp.Usage.InputTokensDetails.CacheWriteTokens),
			ReasoningTokens: int64(resp.Usage.OutputTokensDetails.ReasoningTokens),
		},
		ModelName: p.BaseAgentProvider.Context().ModelName,
	}, nil
}

func (p OpenAIProvider) StreamText(ctx context.Context, params providers.AgentProviderPromptMessageParams, channel chan models.Part) {
	stream := p.response.NewStreaming(ctx, responses.ResponseNewParams{
		Model: p.BaseAgentProvider.Context().ModelName,
		Input: convertInputFromParams(params),
	})

	channel <- models.NewStepStartPart(p.Context(), "streaming started")

	for stream.Next() {
		chunk := stream.Current()
		log.Println("Received chunk:", chunk.Type)
		switch chunk.Type {
		case "response.output_text.delta":
			channel <- models.NewTextPart(p.Context(), string(chunk.Delta))
		case "response.completed":
			channel <- models.NewStepEndPart(p.Context(), "stream completed", models.NewLanguageModelUsage(
				int64(chunk.Response.Usage.OutputTokens),
				int64(chunk.Response.Usage.InputTokens),
				int64(chunk.Response.Usage.InputTokensDetails.CachedTokens)+int64(chunk.Response.Usage.InputTokensDetails.CacheWriteTokens),
				int64(chunk.Response.Usage.OutputTokensDetails.ReasoningTokens),
			))
		}
	}
}

func (p OpenAIProvider) ResolveToolCall(ctx context.Context, params providers.AgentProviderPromptMessageParams, toolParams []models.BaseTool) ([]models.ToolCall, error) {
	response, err := p.response.New(ctx, responses.ResponseNewParams{
		Model: p.BaseAgentProvider.Context().ModelName,
		Input: convertInputFromParams(params),
		Tools: convertToolParamsToInput(toolParams),
	})
	if err != nil {
		return nil, err
	}

	log.Printf("Response ID: %s", response.ID)
	toolCalls := convertOutputToToolCalls(response)

	return toolCalls, nil
}

func NewOpenAIProvider(apiKey, modelName string) OpenAIProvider {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	return newOpenAIProviderWithClient(
		modelName,
		&client.Responses,
	)
}

func newOpenAIProviderWithClient(modelName string, responses openAIResponsesService) OpenAIProvider {
	return OpenAIProvider{
		response: responses,
		BaseAgentProvider: providers.NewBaseAgentProvider(
			models.LanguageModelContext{ModelName: modelName},
		),
	}
}
