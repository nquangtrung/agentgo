package openai

import (
	"context"
	"testing"

	"github.com/openai/openai-go/v3/responses"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/providers/openai/mocks"
)

func TestConvertMessageObjectToInput(t *testing.T) {
	messages := []models.Message{
		models.NewSystemStringMessage("You are a helpful assistant."),
		models.NewHumanStringMessage("Hello, how are you?"),
		models.NewAssistantStringMessage("I'm doing well, thank you! How can I assist you today?"),
	}

	provider := NewOpenAIProvider("mock-api-key", "mock-gpt-5.5")

	input := provider.ConvertMessageObjectToInput(messages)

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

func TestGetInputFromParams(t *testing.T) {
	messages := []models.Message{
		models.NewSystemStringMessage("You are a helpful assistant."),
		models.NewHumanStringMessage("Hello, how are you?"),
		models.NewAssistantStringMessage("I'm doing well, thank you! How can I assist you today?"),
	}

	provider := NewOpenAIProvider("mock-api-key", "mock-gpt-5.5")
	params := providers.AgentProviderPromptMessageParams{
		Messages: messages,
	}

	input := provider.GetInputFromParams(params)

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

func TestGenerateText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockResponsesService := mocks.NewMockopenAIResponsesService(ctrl)
	provider := newOpenAIProviderWithClient(
		"mock-gpt-5.5",
		mockResponsesService,
	)

	mockedResponse := &responses.Response{
		ID: "mocked-response-id",
		Output: []responses.ResponseOutputItemUnion{
			{
				Content: []responses.ResponseOutputMessageContentUnion{
					{
						Text: "Great, thank you!",
						Type: "output_text",
					},
				},
			},
		},
		Usage: responses.ResponseUsage{
			InputTokens: 125,
			InputTokensDetails: responses.ResponseUsageInputTokensDetails{
				CacheWriteTokens: 20,
				CachedTokens:     30,
			},
			OutputTokens: 256,
			OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{
				ReasoningTokens: 50,
			},
			TotalTokens: 125 + 256,
		},
	}
	mockResponsesService.EXPECT().New(
		gomock.Cond(func(context context.Context) bool {
			return true
		}),
		gomock.Cond(func(params responses.ResponseNewParams) bool {
			return params.Model == "mock-gpt-5.5" &&
				len(params.Input.OfInputItemList) == 2 &&
				params.Input.OfInputItemList[0].OfMessage.Role == responses.EasyInputMessageRoleSystem &&
				params.Input.OfInputItemList[1].OfMessage.Role == responses.EasyInputMessageRoleUser
		}),
	).Return(mockedResponse, nil)

	output, err := provider.GenerateText(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{
			models.NewSystemStringMessage("You are a helpful assistant."),
			models.NewHumanStringMessage("Hello, how are you?"),
		},
	})

	assert.Nil(t, err, "should not cause error")
	assert.Equal(t, "Great, thank you!", output.Text, "should have correct response text")
	assert.Equal(t, mockedResponse.Usage.InputTokens, output.Usage.InputTokens, "should have correct input tokens")
	assert.Equal(t, mockedResponse.Usage.OutputTokens, output.Usage.OutputTokens, "should have correct output tokens")
	assert.Equal(t, mockedResponse.Usage.InputTokensDetails.CacheWriteTokens+mockedResponse.Usage.InputTokensDetails.CachedTokens, output.Usage.CachedTokens, "should have correct cached tokens")
	assert.Equal(t, mockedResponse.Usage.OutputTokensDetails.ReasoningTokens, output.Usage.ReasoningTokens, "should have correct reasoning tokens")
}
