package openai

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

	params := providers.AgentProviderPromptMessageParams{
		Messages: messages,
	}

	input := convertInputFromParams(params)

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

func TestConvertMessageObjectToInput(t *testing.T) {
	messages := []models.Message{
		models.NewSystemStringMessage("You are a helpful assistant."),
		models.NewHumanStringMessage("Hello, how are you?"),
		models.NewAssistantStringMessage("I'm doing well, thank you! How can I assist you today?"),
	}

	input := convertMessageObjectToInput(messages)

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
