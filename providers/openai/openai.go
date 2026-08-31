package openai

import (
	"context"
	"log"
	"strings"

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

	usage := models.LanguageModelUsage{
		InputTokens: int64(resp.Usage.InputTokens),
		InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
			CachedTokens:     int64(resp.Usage.InputTokensDetails.CachedTokens),
			CacheWriteTokens: int64(resp.Usage.InputTokensDetails.CacheWriteTokens),
		},
		OutputTokens: int64(resp.Usage.OutputTokens),
		OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
			ReasoningTokens: int64(resp.Usage.OutputTokensDetails.ReasoningTokens),
		},
		TotalTokens: int64(resp.Usage.TotalTokens),
	}
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.InputTokens + usage.OutputTokens
	}

	return models.LanguageModelOutput{
		Text:      resp.OutputText(),
		Usage:     usage,
		ModelName: p.BaseAgentProvider.Context().ModelName,
	}, nil
}

func (p OpenAIProvider) StreamText(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) (models.LanguageModelOutput, error) {
	stream := p.response.NewStreaming(ctx, responses.ResponseNewParams{
		Model: p.BaseAgentProvider.Context().ModelName,
		Input: convertInputFromParams(params),
	})

	emitter.Emit(models.NewStepStartPart(p.Context(), "streaming started"))

	builder := strings.Builder{}
	usage := models.LanguageModelUsage{}
	for stream.Next() {
		chunk := stream.Current()
		log.Println("Received chunk:", chunk.Type)
		switch chunk.Type {
		case "response.output_text.delta":
			emitter.Emit(models.NewTextPart(p.Context(), string(chunk.Delta)))
			builder.Write([]byte(chunk.Delta))
		case "response.completed":
			usage = models.LanguageModelUsage{
				InputTokens: int64(chunk.Response.Usage.InputTokens),
				InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
					CachedTokens:     int64(chunk.Response.Usage.InputTokensDetails.CachedTokens),
					CacheWriteTokens: int64(chunk.Response.Usage.InputTokensDetails.CacheWriteTokens),
				},
				OutputTokens: int64(chunk.Response.Usage.OutputTokens),
				OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
					ReasoningTokens: int64(chunk.Response.Usage.OutputTokensDetails.ReasoningTokens),
				},
				TotalTokens: int64(chunk.Response.Usage.TotalTokens),
			}
			if usage.TotalTokens == 0 {
				usage.TotalTokens = usage.InputTokens + usage.OutputTokens
			}
			emitter.Emit(models.NewStepEndPart(p.Context(), "stream completed", usage))
		}
	}

	return models.LanguageModelOutput{
		Text:      builder.String(),
		Usage:     usage,
		ModelName: p.BaseAgentProvider.Context().ModelName,
	}, nil
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
