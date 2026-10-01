package claude

import (
	"encoding/json"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/utils"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

// convertUsageToLanguageModelUsage maps Chat Completions usage onto the SDK
// usage type. Anthropic's compatibility layer documents both
// prompt_tokens_details and completion_tokens_details as always empty, so
// cached and reasoning tokens stay at zero.
func convertUsageToLanguageModelUsage(usage openai.CompletionUsage) models.LanguageModelUsage {
	return models.LanguageModelUsage{
		InputTokens:         usage.PromptTokens,
		OutputTokens:        usage.CompletionTokens,
		InputTokensDetails:  models.LanguageModelUsageInputTokensDetails{},
		OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{},
		TotalTokens:         usage.TotalTokens,
	}
}

func convertInputFromParams(params providers.AgentProviderPromptMessageParams) []openai.ChatCompletionMessageParamUnion {
	if len(params.Messages) == 0 {
		panic("no messages provided to convertInputFromParams")
	}

	return convertMessageObjectToMessages(params.Messages)
}

// convertMessageObjectToMessages maps SDK messages onto Chat Completions roles.
// Anthropic hoists system messages to the front of the conversation and
// concatenates them, which matches how this SDK already orders them.
func convertMessageObjectToMessages(messages []models.Message) []openai.ChatCompletionMessageParamUnion {
	return utils.Map(
		messages,
		func(message models.Message) openai.ChatCompletionMessageParamUnion {
			switch message.Type() {
			case models.MessageRoleSystem:
				return openai.SystemMessage(message.Content().Text())
			case models.MessageRoleHuman:
				return openai.UserMessage(message.Content().Text())
			case models.MessageRoleAssistant:
				return openai.AssistantMessage(message.Content().Text())
			default:
				return openai.UserMessage(message.Content().Text())
			}
		},
	)
}

// convertResponseFormat rejects structured output instead of silently dropping
// it. Anthropic's OpenAI compatibility layer documents response_format as
// ignored, so sending one would produce unvalidated prose and surface as a
// confusing parse error deep inside the object-repair loop. Failing loudly here
// keeps that distinction visible to the caller.
func convertResponseFormat(format *models.ResponseFormat) (openai.ChatCompletionNewParamsResponseFormatUnion, error) {
	if format == nil {
		return openai.ChatCompletionNewParamsResponseFormatUnion{}, nil
	}

	return openai.ChatCompletionNewParamsResponseFormatUnion{}, &models.StructuredOutputUnsupportedError{
		ProviderName: "claude",
		Reason:       "Anthropic's OpenAI compatibility layer ignores response_format; use the native Anthropic API for structured outputs",
	}
}

// convertToolParamsToInput builds function tools. Anthropic's compatibility
// layer documents `strict` as ignored, so the generated tool JSON is not
// guaranteed to match the supplied input schema.
func convertToolParamsToInput(tools []models.BaseTool) []openai.ChatCompletionToolUnionParam {
	return utils.Map(
		tools,
		func(tool models.BaseTool) openai.ChatCompletionToolUnionParam {
			return openai.ChatCompletionToolUnionParam{
				OfFunction: &openai.ChatCompletionFunctionToolParam{
					Function: shared.FunctionDefinitionParam{
						Name:        tool.Name(),
						Description: openai.String(tool.Description()),
						Parameters:  tool.InputSchema(),
						Strict:      openai.Bool(false),
					},
				},
			}
		},
	)
}

// convertOutputToToolCalls extracts function calls from the first choice. Only
// `function` tool calls are considered; custom tool calls have no equivalent in
// the SDK's tool model.
func convertOutputToToolCalls(completion *openai.ChatCompletion) []models.ToolCall {
	if len(completion.Choices) == 0 {
		return nil
	}

	var toolCalls []models.ToolCall = utils.Map(
		completion.Choices[0].Message.ToolCalls,
		func(call openai.ChatCompletionMessageToolCallUnion) models.ToolCall {
			params := make(map[string]any)
			json.Unmarshal([]byte(call.Function.Arguments), &params)

			return models.ToolCall{
				ToolName: call.Function.Name,
				Params:   params,
				ID:       call.ID,
			}
		},
	)

	return toolCalls
}
