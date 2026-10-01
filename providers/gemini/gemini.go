package gemini

import (
	"context"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

// defaultBaseURL is Google's OpenAI compatibility endpoint. Gemini is spoken to
// over Chat Completions here, not the Responses API.
const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta/openai/"

// streamIncludeUsage requests a final usage-bearing chunk. Without it the stream
// carries no token counts at all.
var streamIncludeUsage = param.NewOpt(true)

type GeminiProvider struct {
	providers.BaseAgentProvider
	chat geminiChatService
}

func (p GeminiProvider) GenerateText(
	ctx context.Context,
	params providers.AgentProviderPromptMessageParams,
) (models.LanguageModelOutput, error) {
	completion, err := p.chat.New(ctx, openai.ChatCompletionNewParams{
		Model:          p.BaseAgentProvider.Context().ModelName,
		Messages:       convertInputFromParams(params),
		ResponseFormat: convertResponseFormat(params.ResponseFormat),
	})
	if err != nil {
		return models.LanguageModelOutput{}, err
	}

	if len(completion.Choices) == 0 {
		return models.LanguageModelOutput{}, &models.ExecutionContextError{
			Message: "gemini returned no choices",
		}
	}

	usage := convertUsageToLanguageModelUsage(completion.Usage)
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.InputTokens + usage.OutputTokens
	}

	return models.LanguageModelOutput{
		Text:      completion.Choices[0].Message.Content,
		Usage:     usage,
		ModelName: p.BaseAgentProvider.Context().ModelName,
	}, nil
}

func (p GeminiProvider) StreamText(
	ctx context.Context,
	params providers.AgentProviderPromptMessageParams,
	emitter models.PartEmitter) (models.LanguageModelOutput, error) {
	stream := p.chat.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:          p.BaseAgentProvider.Context().ModelName,
		Messages:       convertInputFromParams(params),
		ResponseFormat: convertResponseFormat(params.ResponseFormat),
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: streamIncludeUsage,
		},
	})

	emitter.Emit(models.NewStepStartPart(p.Context(), "streaming started"))

	builder := strings.Builder{}
	usage := models.LanguageModelUsage{}
	for stream.Next() {
		chunk := stream.Current()

		// The usage-bearing chunk carries no choices, so it is handled first.
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				emitter.Emit(models.NewTextPart(p.Context(), delta))
				builder.WriteString(delta)
			}
		}

		if chunk.Usage.TotalTokens > 0 || chunk.Usage.PromptTokens > 0 || chunk.Usage.CompletionTokens > 0 {
			usage = convertUsageToLanguageModelUsage(chunk.Usage)
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

func (p GeminiProvider) ResolveToolCall(
	ctx context.Context,
	params providers.AgentProviderPromptMessageParams,
	toolParams []models.BaseTool) (models.LanguageModelToolCallResolveOutput, error) {
	completion, err := p.chat.New(ctx, openai.ChatCompletionNewParams{
		Model:    p.BaseAgentProvider.Context().ModelName,
		Messages: convertInputFromParams(params),
		Tools:    convertToolParamsToInput(toolParams),
	})
	if err != nil {
		return models.LanguageModelToolCallResolveOutput{}, err
	}

	return models.LanguageModelToolCallResolveOutput{
		ToolCalls: convertOutputToToolCalls(completion),
		Usage:     convertUsageToLanguageModelUsage(completion.Usage),
		ModelName: p.BaseAgentProvider.Context().ModelName,
		Text:      firstChoiceContent(completion),
	}, nil
}

// NewGeminiProvider builds a provider backed by Google's compatibility endpoint.
func NewGeminiProvider(apiKey, modelName string) GeminiProvider {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(defaultBaseURL),
	)

	return newGeminiProviderWithClient(
		modelName,
		&client.Chat.Completions,
	)
}

func newGeminiProviderWithClient(modelName string, chat geminiChatService) GeminiProvider {
	return GeminiProvider{
		chat: chat,
		BaseAgentProvider: providers.NewBaseAgentProvider(
			models.LanguageModelContext{ModelName: modelName},
		),
	}
}
