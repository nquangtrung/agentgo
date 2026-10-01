package openai

import (
	"encoding/json"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/utils"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func convertInputFromParams(params providers.AgentProviderPromptMessageParams) responses.ResponseNewParamsInputUnion {
	if len(params.Messages) == 0 {
		panic("no messages provided to GetInputFromParams")
	}

	return convertMessageObjectToInput(params.Messages)
}

// convertMessageObjectToInput maps SDK messages onto Responses API input items.
//
// Tool calls and tool results become dedicated function_call / function_call_output
// items rather than message items, which is what lets the API correlate a result
// with the call that produced it. The Responses API accepts one input item per
// tool part, so a message carrying several parts expands to several items.
func convertMessageObjectToInput(messages []models.Message) responses.ResponseNewParamsInputUnion {
	var inputItems []responses.ResponseInputItemUnionParam

	for _, message := range messages {
		content := message.Content()
		toolCalls := content.ToolCalls()
		toolResults := content.ToolResults()

		// Tool-bearing messages map to dedicated item types, not message items.
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

			// Any narration alongside the tool parts is a normal output message.
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

// convertTextMessageToInputItem maps a text-only message. A tool call without an
// id cannot be expressed as a function_call item, so it degrades to prose rather
// than being sent as an uncorrelatable item.
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
		// Human, tool results without ids, and any unknown role.
		return responses.ResponseInputItemParamOfMessage(
			responses.ResponseInputMessageContentListParam{
				responses.ResponseInputContentParamOfInputText(text),
			},
			responses.EasyInputMessageRoleUser,
		)
	}
}

// convertResponseFormat maps a models.ResponseFormat to the OpenAI Responses
// API text configuration. A nil format yields the zero value, which the API
// treats as plain text.
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

func convertToolParamsToInput(tools []models.BaseTool) []responses.ToolUnionParam {
	return utils.Map(tools, func(tool models.BaseTool) responses.ToolUnionParam {
		openAiTool := responses.ToolParamOfFunction(tool.Name(), tool.InputSchema(), true)
		openAiTool.OfFunction.Description = openai.String(tool.Description())
		return openAiTool
	})
}

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
