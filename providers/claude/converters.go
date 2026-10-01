package claude

import (
	"encoding/json"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/utils"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
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
//
// Tool calls ride on the assistant turn's tool_calls array and tool results
// become role:"tool" messages carrying tool_call_id. Anthropic's native
// Messages API requires a tool_result to reference a preceding tool_use by id
// and errors otherwise, so replaying both is what keeps tool calling valid
// against the compatibility layer and would remain valid natively.
func convertMessageObjectToMessages(messages []models.Message) []openai.ChatCompletionMessageParamUnion {
	var converted []openai.ChatCompletionMessageParamUnion

	for _, message := range messages {
		content := message.Content()
		toolCalls := content.ToolCalls()
		toolResults := content.ToolResults()

		switch {
		case len(toolCalls) > 0:
			converted = append(converted, convertToolCallMessage(toolCalls, content.Text()))
			for _, result := range toolResults {
				converted = append(converted, openai.ToolMessage(result.Payload(), result.ToolCallID()))
			}
		case len(toolResults) > 0:
			for _, result := range toolResults {
				converted = append(converted, openai.ToolMessage(result.Payload(), result.ToolCallID()))
			}
		default:
			converted = append(converted, convertTextMessage(message))
		}
	}

	return converted
}

// convertToolCallMessage builds the assistant turn replaying tool calls. Calls
// without an id are skipped, since an uncorrelatable tool_use would leave its
// result with nothing to reference.
func convertToolCallMessage(toolCalls []models.ToolCallContentPart, text string) openai.ChatCompletionMessageParamUnion {
	var assistant openai.ChatCompletionAssistantMessageParam
	if text != "" {
		assistant.Content.OfString = param.NewOpt(text)
	}

	for _, call := range toolCalls {
		if call.ID() == "" {
			continue
		}
		arguments := utils.Must(json.Marshal(call.Input()))
		assistant.ToolCalls = append(assistant.ToolCalls, openai.ChatCompletionMessageToolCallUnionParam{
			OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
				ID: call.ID(),
				Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
					Name:      call.Name(),
					Arguments: string(arguments),
				},
			},
		})
	}

	return openai.ChatCompletionMessageParamUnion{OfAssistant: &assistant}
}

func convertTextMessage(message models.Message) openai.ChatCompletionMessageParamUnion {
	text := message.Content().Text()

	switch message.Type() {
	case models.MessageRoleSystem:
		return openai.SystemMessage(text)
	case models.MessageRoleAssistant:
		return openai.AssistantMessage(text)
	default:
		return openai.UserMessage(text)
	}
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

// firstChoiceContent returns the first choice's text, or "" when the completion
// carries no choices.
func firstChoiceContent(completion *openai.ChatCompletion) string {
	if len(completion.Choices) == 0 {
		return ""
	}
	return completion.Choices[0].Message.Content
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
