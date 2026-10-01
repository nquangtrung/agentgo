package deepseek

import (
	"encoding/json"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/utils"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// convertUsageToLanguageModelUsage maps the Responses API usage object onto the
// SDK usage type. DeepSeek reports context-cache hits under
// input_tokens_details.cached_tokens and chain-of-thought tokens under
// output_tokens_details.reasoning_tokens. It does not report cache writes, so
// CacheWriteTokens is always left at zero.
func convertUsageToLanguageModelUsage(usage responses.ResponseUsage) models.LanguageModelUsage {
	return models.LanguageModelUsage{
		InputTokens: int64(usage.InputTokens),
		InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
			CachedTokens:     int64(usage.InputTokensDetails.CachedTokens),
			CacheWriteTokens: int64(usage.InputTokensDetails.CacheWriteTokens),
		},
		OutputTokens: int64(usage.OutputTokens),
		OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
			ReasoningTokens: int64(usage.OutputTokensDetails.ReasoningTokens),
		},
		TotalTokens: int64(usage.TotalTokens),
	}
}

func convertInputFromParams(params providers.AgentProviderPromptMessageParams) responses.ResponseNewParamsInputUnion {
	if len(params.Messages) == 0 {
		panic("no messages provided to convertInputFromParams")
	}

	return convertMessageObjectToInput(params.Messages)
}

// convertMessageObjectToInput maps SDK messages onto Responses API input items.
//
// Tool calls and tool results become dedicated function_call / function_call_output
// items so the API can correlate a result with the call that produced it. A
// message with several tool parts expands into several items.
func convertMessageObjectToInput(messages []models.Message) responses.ResponseNewParamsInputUnion {
	var inputItems []responses.ResponseInputItemUnionParam

	for _, message := range messages {
		content := message.Content()
		toolCalls := content.ToolCalls()
		toolResults := content.ToolResults()

		if len(toolCalls) > 0 || len(toolResults) > 0 {
			for _, call := range toolCalls {
				arguments := utils.Must(json.Marshal(call.Input()))
				inputItems = append(inputItems, responses.ResponseInputItemParamOfFunctionCall(
					string(arguments),
					call.ID(),
					call.Name(),
				))
			}
			for _, result := range toolResults {
				inputItems = append(inputItems, responses.ResponseInputItemParamOfFunctionCallOutput(
					result.ToolCallID(),
					result.Payload(),
				))
			}
			if text := content.Text(); text != "" {
				inputItems = append(inputItems, responses.ResponseInputItemParamOfOutputMessage(
					[]responses.ResponseOutputMessageContentUnionParam{
						{OfOutputText: &responses.ResponseOutputTextParam{Text: text}},
					},
					"",
					responses.ResponseOutputMessageStatusCompleted,
				))
			}
			continue
		}

		inputItems = append(inputItems, convertTextMessageToInputItem(message))
	}

	return responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems}
}

func convertTextMessageToInputItem(message models.Message) responses.ResponseInputItemUnionParam {
	text := message.Content().Text()

	switch message.Type() {
	case models.MessageRoleSystem:
		return responses.ResponseInputItemParamOfMessage(
			responses.ResponseInputMessageContentListParam{
				responses.ResponseInputContentParamOfInputText(text),
			},
			responses.EasyInputMessageRoleSystem,
		)
	case models.MessageRoleAssistant:
		return responses.ResponseInputItemParamOfOutputMessage(
			[]responses.ResponseOutputMessageContentUnionParam{
				{OfOutputText: &responses.ResponseOutputTextParam{Text: text}},
			},
			"",
			responses.ResponseOutputMessageStatusCompleted,
		)
	default:
		return responses.ResponseInputItemParamOfMessage(
			responses.ResponseInputMessageContentListParam{
				responses.ResponseInputContentParamOfInputText(text),
			},
			responses.EasyInputMessageRoleUser,
		)
	}
}

// convertResponseFormat maps a models.ResponseFormat to the Responses API text
// configuration. DeepSeek fully supports the `text` parameter including
// `format`, so JSON-schema structured output works the same way it does for
// OpenAI. A nil format yields the zero value, which the API treats as plain
// text.
func convertResponseFormat(format *models.ResponseFormat) responses.ResponseTextConfigParam {
	if format == nil {
		return responses.ResponseTextConfigParam{}
	}

	jsonSchema := &responses.ResponseFormatTextJSONSchemaConfigParam{
		Name:   format.Name,
		Schema: format.JSONSchema,
		Strict: openai.Bool(format.Strict),
	}
	if format.Description != "" {
		jsonSchema.Description = openai.String(format.Description)
	}

	return responses.ResponseTextConfigParam{
		Format: responses.ResponseFormatTextConfigUnionParam{OfJSONSchema: jsonSchema},
	}
}

// convertToolParamsToInput builds function tools. DeepSeek supports `function`
// tools and ignores every other tool type.
func convertToolParamsToInput(tools []models.BaseTool) []responses.ToolUnionParam {
	return utils.Map(tools, func(tool models.BaseTool) responses.ToolUnionParam {
		deepSeekTool := responses.ToolParamOfFunction(tool.Name(), tool.InputSchema(), true)
		deepSeekTool.OfFunction.Description = openai.String(tool.Description())
		return deepSeekTool
	})
}

// convertOutputToToolCalls extracts function calls from a completed response.
// Responses that only contain text yield a nil slice.
func convertOutputToToolCalls(response *responses.Response) []models.ToolCall {
	filtered := utils.Filter(
		response.Output,
		func(outputItem responses.ResponseOutputItemUnion) bool {
			return outputItem.Type == "function_call"
		})

	var toolCalls []models.ToolCall = utils.Map(
		filtered,
		func(outputItem responses.ResponseOutputItemUnion) models.ToolCall {
			call := outputItem.AsFunctionCall()
			params := make(map[string]any)
			json.Unmarshal(
				[]byte(
					utils.Ternary(
						call.Arguments != "",
						call.Arguments,
						outputItem.Arguments.OfString,
					),
				),
				&params,
			)
			return models.ToolCall{
				ToolName: utils.Ternary(
					call.Name != "",
					call.Name,
					outputItem.Name,
				),
				Params: params,
				ID:     call.CallID,
			}
		},
	)

	return toolCalls
}
