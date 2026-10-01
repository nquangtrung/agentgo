package deepseek

import (
	"context"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

// defaultBaseURL is the DeepSeek-hosted, OpenAI-compatible endpoint. The
// Responses API, streaming, function tools and `text.format` structured output
// are all served from this single base URL.
const defaultBaseURL = "https://api.deepseek.com"

type DeepSeekProvider struct {
	providers.BaseAgentProvider
	response deepSeekResponsesService
}

func (p DeepSeekProvider) GenerateText(
	ctx context.Context,
	params providers.AgentProviderPromptMessageParams,
) (models.LanguageModelOutput, error) {
	resp, err := p.response.New(ctx, responses.ResponseNewParams{
		Model: p.BaseAgentProvider.Context().ModelName,
		Input: convertInputFromParams(params),
		Text:  convertResponseFormat(params.ResponseFormat),
	})
	if err != nil {
		return models.LanguageModelOutput{}, err
	}

	usage := convertUsageToLanguageModelUsage(resp.Usage)
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.InputTokens + usage.OutputTokens
	}

	return models.LanguageModelOutput{
		Text:      resp.OutputText(),
		Usage:     usage,
		ModelName: p.BaseAgentProvider.Context().ModelName,
	}, nil
}

// StreamText consumes DeepSeek's semantic SSE stream. DeepSeek terminates the
// stream with a `response.completed` event rather than the OpenAI `[DONE]`
// sentinel, so termination is detected from the event itself.
func (p DeepSeekProvider) StreamText(
	ctx context.Context,
	params providers.AgentProviderPromptMessageParams,
	emitter models.PartEmitter) (models.LanguageModelOutput, error) {
	stream := p.response.NewStreaming(ctx, responses.ResponseNewParams{
		Model: p.BaseAgentProvider.Context().ModelName,
		Input: convertInputFromParams(params),
		Text:  convertResponseFormat(params.ResponseFormat),
	})

	emitter.Emit(models.NewStepStartPart(p.Context(), "streaming started"))

	builder := strings.Builder{}
	usage := models.LanguageModelUsage{}
	for stream.Next() {
		chunk := stream.Current()
		switch chunk.Type {
		case "response.output_text.delta":
			emitter.Emit(models.NewTextPart(p.Context(), string(chunk.Delta)))
			builder.Write([]byte(chunk.Delta))
		case "response.completed":
			usage = convertUsageToLanguageModelUsage(chunk.Response.Usage)
			if usage.TotalTokens == 0 {
				usage.TotalTokens = usage.InputTokens + usage.OutputTokens
			}
			emitter.Emit(models.NewStepEndPart(p.Context(), "stream completed", usage))
		}
	}
	if err := stream.Err(); err != nil {
		return models.LanguageModelOutput{}, err
	}

	return models.LanguageModelOutput{
		Text:      builder.String(),
		Usage:     usage,
		ModelName: p.BaseAgentProvider.Context().ModelName,
	}, nil
}

func (p DeepSeekProvider) ResolveToolCall(
	ctx context.Context,
	params providers.AgentProviderPromptMessageParams,
	toolParams []models.BaseTool) (models.LanguageModelToolCallResolveOutput, error) {
	response, err := p.response.New(ctx, responses.ResponseNewParams{
		Model: p.BaseAgentProvider.Context().ModelName,
		Input: convertInputFromParams(params),
		Tools: convertToolParamsToInput(toolParams),
	})
	if err != nil {
		return models.LanguageModelToolCallResolveOutput{}, err
	}

	return models.LanguageModelToolCallResolveOutput{
		ToolCalls: convertOutputToToolCalls(response),
		Usage:     convertUsageToLanguageModelUsage(response.Usage),
		ModelName: p.BaseAgentProvider.Context().ModelName,
	}, nil
}

// NewDeepSeekProvider builds a provider backed by the real DeepSeek endpoint.
func NewDeepSeekProvider(apiKey, modelName string) DeepSeekProvider {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(defaultBaseURL),
	)

	return newDeepSeekProviderWithClient(
		modelName,
		&client.Responses,
	)
}

func newDeepSeekProviderWithClient(modelName string, responses deepSeekResponsesService) DeepSeekProvider {
	return DeepSeekProvider{
		response: responses,
		BaseAgentProvider: providers.NewBaseAgentProvider(
			models.LanguageModelContext{ModelName: modelName},
		),
	}
}
