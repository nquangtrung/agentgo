package gemini

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
// usage type. Gemini reports cached prompt tokens under
// prompt_tokens_details.cached_tokens and thinking tokens under
// completion_tokens_details.reasoning_tokens.
func convertUsageToLanguageModelUsage(usage openai.CompletionUsage) models.LanguageModelUsage {
	return models.LanguageModelUsage{
		InputTokens: usage.PromptTokens,
		InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
			CachedTokens: usage.PromptTokensDetails.CachedTokens,
		},
		OutputTokens: usage.CompletionTokens,
		OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
			ReasoningTokens: usage.CompletionTokensDetails.ReasoningTokens,
		},
		TotalTokens: usage.TotalTokens,
	}
}

func convertInputFromParams(params providers.AgentProviderPromptMessageParams) []openai.ChatCompletionMessageParamUnion {
	if len(params.Messages) == 0 {
		panic("no messages provided to convertInputFromParams")
	}

	return convertMessageObjectToMessages(params.Messages)
}

// convertMessageObjectToMessages maps SDK messages onto Chat Completions roles.
//
// Tool calls ride on the assistant turn's tool_calls array and tool results
// become role:"tool" messages carrying tool_call_id, which is how the Chat
// Completions shape correlates the two. A single tool result message maps to
// exactly one tool message, matching that wire format's one-call-per-message
// rule.
func convertMessageObjectToMessages(messages []models.Message) []openai.ChatCompletionMessageParamUnion {
	var converted []openai.ChatCompletionMessageParamUnion

	for _, message := range messages {
		content := message.Content()
		toolCalls := content.ToolCalls()
		toolResults := content.ToolResults()

		switch {
		case len(toolCalls) > 0:
			converted = append(converted, convertToolCallMessage(toolCalls, content.Text()))
			// Fall through: a tool result in the same message is unusual, but
			// emitting it keeps nothing silently dropped.
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
// without an id are skipped: an uncorrelatable tool_call would leave the
// matching result dangling, which the API rejects.
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

// convertResponseFormat maps a models.ResponseFormat to the Chat Completions
// response format. Gemini supports the json_schema variant, including
// `strict` and `additional_properties`, though fully recursive schemas are not
// supported. A nil format yields the zero value, which the API treats as plain
// text.
func convertResponseFormat(format *models.ResponseFormat) openai.ChatCompletionNewParamsResponseFormatUnion {
	if format == nil {
		return openai.ChatCompletionNewParamsResponseFormatUnion{}
	}

	jsonSchema := &shared.ResponseFormatJSONSchemaParam{
		JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:   format.Name,
			Schema: format.JSONSchema,
			Strict: openai.Bool(format.Strict),
		},
	}
	if format.Description != "" {
		jsonSchema.JSONSchema.Description = openai.String(format.Description)
	}

	return openai.ChatCompletionNewParamsResponseFormatUnion{OfJSONSchema: jsonSchema}
}

// convertToolParamsToInput builds function tools. Gemini's function `parameters`
// field follows the OpenAPI spec rather than JSON Schema, which differs on a
// few keywords; the schema is passed through as-is.
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

// convertOutputToToolCalls extracts function calls from the first choice.
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
