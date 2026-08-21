package openai

import (
	"encoding/json"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

func convertInputFromParams(params providers.AgentProviderPromptMessageParams) responses.ResponseNewParamsInputUnion {
	if len(params.Messages) == 0 {
		panic("no messages provided to GetInputFromParams")
	}

	return convertMessageObjectToInput(params.Messages)
}

func convertMessageObjectToInput(messages []models.Message) responses.ResponseNewParamsInputUnion {
	var inputItems []responses.ResponseInputItemUnionParam = utils.Map(
		messages,
		func(message models.Message) responses.ResponseInputItemUnionParam {
			switch message.Type() {
			case models.MessageRoleSystem:
				return responses.ResponseInputItemParamOfMessage(
					responses.ResponseInputMessageContentListParam{
						responses.ResponseInputContentParamOfInputText(message.Content().Text()),
					},
					responses.EasyInputMessageRoleSystem,
				)
			case models.MessageRoleHuman:
				return responses.ResponseInputItemParamOfMessage(
					responses.ResponseInputMessageContentListParam{
						responses.ResponseInputContentParamOfInputText(message.Content().Text()),
					},
					responses.EasyInputMessageRoleUser,
				)
			case models.MessageRoleAssistant:
				return responses.ResponseInputItemParamOfOutputMessage(
					[]responses.ResponseOutputMessageContentUnionParam{
						{OfOutputText: &responses.ResponseOutputTextParam{
							Text: message.Content().Text(),
							// Annotations: []responses.ResponseOutputTextAnnotationUnionParam{},
						}},
					},
					"",
					responses.ResponseOutputMessageStatusCompleted,
				)
			default:
				return responses.ResponseInputItemParamOfMessage(
					responses.ResponseInputMessageContentListParam{
						responses.ResponseInputContentParamOfInputText(message.Content().Text()),
					},
					responses.EasyInputMessageRoleUser,
				)
			}
		},
	)

	return responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems}
}

func convertToolParamsToInput(tools []models.BaseTool) []responses.ToolUnionParam {
	return utils.Map(tools, func(tool models.BaseTool) responses.ToolUnionParam {
		// parameters := map[string]any{
		// 	"type": "object",
		// 	"properties": map[string]any{
		// 		"location": map[string]any{"type": "string", "description": "Return the temperature for this location."},
		// 	},
		// 	"required":             []string{"location"},
		// 	"additionalProperties": false,
		// }

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
				[]byte(utils.Ternary(call.Arguments != "", call.Arguments, outputItem.Arguments.OfString)),
				&params,
			)
			return models.ToolCall{
				ToolName: utils.Ternary(call.Name != "", call.Name, outputItem.Name),
				Params:   params,
			}
		},
	)

	return toolCalls
}
