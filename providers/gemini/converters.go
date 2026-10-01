package gemini

import (
	"encoding/json"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/utils"
	"github.com/openai/openai-go/v3"
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
