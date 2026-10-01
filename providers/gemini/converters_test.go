package gemini

import (
	"testing"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/openai/openai-go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertInputFromParams(t *testing.T) {
	messages := []models.Message{
		models.NewSystemStringMessage("You are a helpful assistant."),
		models.NewHumanStringMessage("Hello, how are you?"),
		models.NewAssistantStringMessage("I'm doing well, thank you!"),
	}

	converted := convertInputFromParams(providers.AgentProviderPromptMessageParams{Messages: messages})
	require.Len(t, converted, 3)

	// All three roles carry content as a union wrapping an optional string.
	assert.Equal(t, "You are a helpful assistant.", converted[0].OfSystem.Content.OfString.Value)

	assert.Equal(t, "Hello, how are you?", converted[1].OfUser.Content.OfString.Value)

	assert.Equal(t, "I'm doing well, thank you!", converted[2].OfAssistant.Content.OfString.Value)
}

func TestConvertInputFromParamsPanicsWithoutMessages(t *testing.T) {
	assert.Panics(t, func() {
		convertInputFromParams(providers.AgentProviderPromptMessageParams{})
	})
}

func TestConvertResponseFormatNil(t *testing.T) {
	assert.Equal(t, openai.ChatCompletionNewParamsResponseFormatUnion{}, convertResponseFormat(nil))
}

func TestConvertResponseFormat(t *testing.T) {
	format := convertResponseFormat(&models.ResponseFormat{
		Name:        "recipe",
		Description: "A recipe",
		Strict:      true,
		JSONSchema:  map[string]any{"type": "object"},
	})

	jsonSchema := format.OfJSONSchema
	require.NotNil(t, jsonSchema, "should use the json_schema variant")
	assert.Equal(t, "recipe", jsonSchema.JSONSchema.Name)
	assert.Equal(t, "A recipe", jsonSchema.JSONSchema.Description.Value)
	assert.True(t, jsonSchema.JSONSchema.Strict.Value)

	schema, ok := jsonSchema.JSONSchema.Schema.(map[string]any)
	require.True(t, ok, "schema should round-trip as an object")
	assert.Equal(t, "object", schema["type"])
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
	require.Len(t, converted, 1)
	require.NotNil(t, converted[0].OfFunction)
	assert.Equal(t, "get-weather", converted[0].OfFunction.Function.Name)
	assert.Equal(t, "Get the current weather", converted[0].OfFunction.Function.Description.Value)
	assert.Equal(t, "object", converted[0].OfFunction.Function.Parameters["type"])
	// strict is not requested; Gemini's function parameters are OpenAPI-shaped.
	assert.False(t, converted[0].OfFunction.Function.Strict.Value)
}

func TestConvertOutputToToolCalls(t *testing.T) {
	completion := &openai.ChatCompletion{
		Choices: []openai.ChatCompletionChoice{{
			Message: openai.ChatCompletionMessage{
				ToolCalls: []openai.ChatCompletionMessageToolCallUnion{{
					ID:   "toolu_abc",
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "get-weather",
						Arguments: `{"location":"Bonn"}`,
					},
				}},
			},
		}},
	}

	toolCalls := convertOutputToToolCalls(completion)
	require.Len(t, toolCalls, 1)
	assert.Equal(t, "get-weather", toolCalls[0].ToolName)
	assert.Equal(t, map[string]any{"location": "Bonn"}, toolCalls[0].Params)
	assert.Equal(t, "toolu_abc", toolCalls[0].ID)
}

func TestConvertOutputToToolCallsWithoutChoices(t *testing.T) {
	assert.Nil(t, convertOutputToToolCalls(&openai.ChatCompletion{}))
}

func TestConvertOutputToToolCallsWithoutCalls(t *testing.T) {
	completion := &openai.ChatCompletion{
		Choices: []openai.ChatCompletionChoice{{
			Message: openai.ChatCompletionMessage{Content: "plain text"},
		}},
	}

	assert.Empty(t, convertOutputToToolCalls(completion))
}

func TestConvertUsageToLanguageModelUsage(t *testing.T) {
	usage := convertUsageToLanguageModelUsage(openai.CompletionUsage{
		PromptTokens:            100,
		CompletionTokens:        50,
		TotalTokens:             150,
		PromptTokensDetails:     openai.CompletionUsagePromptTokensDetails{CachedTokens: 64},
		CompletionTokensDetails: openai.CompletionUsageCompletionTokensDetails{ReasoningTokens: 20},
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
