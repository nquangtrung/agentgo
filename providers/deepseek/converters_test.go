package deepseek

import (
	"testing"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/openai/openai-go/v3/responses"
	"github.com/stretchr/testify/assert"
)

func TestConvertInputFromParams(t *testing.T) {
	messages := []models.Message{
		models.NewSystemStringMessage("You are a helpful assistant."),
		models.NewHumanStringMessage("Hello, how are you?"),
		models.NewAssistantStringMessage("I'm doing well, thank you! How can I assist you today?"),
	}

	input := convertInputFromParams(providers.AgentProviderPromptMessageParams{Messages: messages})

	assert.Equal(t, 3, len(input.OfInputItemList), "should have 3 input items")

	systemInput := input.OfInputItemList[0].OfMessage
	assert.Equal(t, responses.EasyInputMessageRoleSystem, systemInput.Role)
	assert.Equal(t, "You are a helpful assistant.", systemInput.Content.OfInputItemContentList[0].OfInputText.Text)

	humanInput := input.OfInputItemList[1].OfMessage
	assert.Equal(t, responses.EasyInputMessageRoleUser, humanInput.Role)
	assert.Equal(t, "Hello, how are you?", humanInput.Content.OfInputItemContentList[0].OfInputText.Text)

	assistantInput := input.OfInputItemList[2].OfOutputMessage
	assert.Equal(t, responses.ResponseOutputMessageStatusCompleted, assistantInput.Status)
	assert.Equal(t, "I'm doing well, thank you! How can I assist you today?", assistantInput.Content[0].OfOutputText.Text)
}

func TestConvertInputFromParamsPanicsWithoutMessages(t *testing.T) {
	assert.Panics(t, func() {
		convertInputFromParams(providers.AgentProviderPromptMessageParams{})
	})
}

func TestConvertResponseFormatNil(t *testing.T) {
	format := convertResponseFormat(nil)
	assert.Equal(t, responses.ResponseTextConfigParam{}, format)
}

func TestConvertResponseFormat(t *testing.T) {
	format := convertResponseFormat(&models.ResponseFormat{
		Name:        "recipe",
		Description: "A recipe",
		Strict:      true,
		JSONSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
			"required": []string{"name"},
		},
	})

	jsonSchema := format.Format.OfJSONSchema
	assert.NotNil(t, jsonSchema, "should use the json_schema variant")
	assert.Equal(t, "recipe", jsonSchema.Name)
	assert.Equal(t, "A recipe", jsonSchema.Description.Value)
	assert.True(t, jsonSchema.Strict.Value, "strict should be set")
	assert.Equal(t, "object", jsonSchema.Schema["type"])
}

func TestConvertToolParamsToInput(t *testing.T) {
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name:        "get-weather",
			Description: "Get the current weather",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"location": map[string]any{"type": "string"}},
				"required":   []string{"location"},
			},
			Fn: func(p models.ToolExecuteParams) models.ToolExecuteOutput { return models.ToolExecuteOutput{} },
		}),
	}

	converted := convertToolParamsToInput(tools)
	assert.Len(t, converted, 1)
	assert.Equal(t, "get-weather", converted[0].OfFunction.Name)
	assert.Equal(t, "Get the current weather", converted[0].OfFunction.Description.Value)
	assert.Equal(t, "object", converted[0].OfFunction.Parameters["type"])
}

func TestConvertOutputToToolCalls(t *testing.T) {
	response := &responses.Response{
		Output: []responses.ResponseOutputItemUnion{
			{Content: []responses.ResponseOutputMessageContentUnion{{Text: "some text", Type: "output_text"}}},
			{
				Type:      "function_call",
				Name:      "get-weather",
				Arguments: responses.ResponseOutputItemUnionArguments{OfString: `{"location":"Bonn"}`},
			},
		},
	}

	toolCalls := convertOutputToToolCalls(response)
	assert.Len(t, toolCalls, 1)
	assert.Equal(t, "get-weather", toolCalls[0].ToolName)
	assert.Equal(t, map[string]any{"location": "Bonn"}, toolCalls[0].Params)
}

func TestConvertOutputToToolCallsWithoutCalls(t *testing.T) {
	response := &responses.Response{
		Output: []responses.ResponseOutputItemUnion{
			{Content: []responses.ResponseOutputMessageContentUnion{{Text: "some text", Type: "output_text"}}},
		},
	}

	assert.Empty(t, convertOutputToToolCalls(response))
}

func TestConvertUsageToLanguageModelUsage(t *testing.T) {
	usage := convertUsageToLanguageModelUsage(responses.ResponseUsage{
		InputTokens:         100,
		OutputTokens:        50,
		InputTokensDetails:  responses.ResponseUsageInputTokensDetails{CachedTokens: 64, CacheWriteTokens: 0},
		OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{ReasoningTokens: 20},
		TotalTokens:         150,
	})

	assert.Equal(t, models.LanguageModelUsage{
		InputTokens: 100,
		InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
			CachedTokens: 64,
		},
		OutputTokens: 50,
		OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
			ReasoningTokens: 20,
		},
		TotalTokens: 150,
	}, usage)
}
