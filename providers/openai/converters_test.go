package openai

import (
	"testing"

	"github.com/openai/openai-go/v3/responses"
	"github.com/stretchr/testify/assert"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
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
