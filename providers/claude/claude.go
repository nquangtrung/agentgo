package claude

import (
	"context"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

// defaultBaseURL is Anthropic's OpenAI SDK compatibility endpoint. Claude is
// spoken to over Chat Completions here, not the Responses API, and the layer
// carries documented gaps: `response_format` and `strict` are ignored, usage
// details are always empty, and system messages are hoisted to the front.
const defaultBaseURL = "https://api.anthropic.com/v1/"

// streamIncludeUsage requests a final usage-bearing chunk. Without it the stream
// carries no token counts at all.
var streamIncludeUsage = param.NewOpt(true)

type ClaudeProvider struct {
	providers.BaseAgentProvider
	chat claudeChatService
}

func (p ClaudeProvider) GenerateText(
	ctx context.Context,
	params providers.AgentProviderPromptMessageParams,
) (models.LanguageModelOutput, error) {
	responseFormat, err := convertResponseFormat(params.ResponseFormat)
	if err != nil {
		return models.LanguageModelOutput{}, err
	}

	completion, err := p.chat.New(ctx, openai.ChatCompletionNewParams{
		Model:          p.BaseAgentProvider.Context().ModelName,
		Messages:       convertInputFromParams(params),
		ResponseFormat: responseFormat,
	})
	if err != nil {
		return models.LanguageModelOutput{}, err
	}

	if len(completion.Choices) == 0 {
		return models.LanguageModelOutput{}, &models.ExecutionContextError{
			Message: "claude returned no choices",
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

func (p ClaudeProvider) StreamText(
	ctx context.Context,
	params providers.AgentProviderPromptMessageParams,
	emitter models.PartEmitter) (models.LanguageModelOutput, error) {
	responseFormat, err := convertResponseFormat(params.ResponseFormat)
	if err != nil {
		return models.LanguageModelOutput{}, err
	}

	stream := p.chat.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:          p.BaseAgentProvider.Context().ModelName,
		Messages:       convertInputFromParams(params),
		ResponseFormat: responseFormat,
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

func (p ClaudeProvider) ResolveToolCall(
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

// NewClaudeProvider builds a provider backed by Anthropic's compatibility
// endpoint.
func NewClaudeProvider(apiKey, modelName string) ClaudeProvider {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(defaultBaseURL),
	)

	return newClaudeProviderWithClient(
		modelName,
		&client.Chat.Completions,
	)
}

func newClaudeProviderWithClient(modelName string, chat claudeChatService) ClaudeProvider {
	return ClaudeProvider{
		chat: chat,
		BaseAgentProvider: providers.NewBaseAgentProvider(
			models.LanguageModelContext{ModelName: modelName},
		),
	}
}
