package tests

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"trontria.com/agentgo"
	"trontria.com/agentgo/mocks"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

func TestGenerateText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Say this is a test"
	modelName := "mocked-llm-3.6-flash"
	result := "This is a test"

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := agentgo.Params{
		Prompt:   prompt,
		Provider: mockProvider,
	}
	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})
	mockProvider.EXPECT().GenerateText(gomock.Cond(
		func(p providers.AgentProviderPromptMessageParams) bool {
			// assert.Equal(t, len(p.Messages), 1, "should")
			if len(p.Messages) != 1 {
				return false
			}

			if p.Messages[0].Content().Text() != params.Prompt {
				return false
			}

			return true
		}),
	).Return(
		models.LanguageModelOutput{
			Text: result,
			Usage: models.LanguageModelUsage{
				OutputTokens:    123,
				InputTokens:     245,
				CachedTokens:    21,
				ReasoningTokens: 12,
			},
			ModelName: modelName,
		},
		nil,
	)

	output, err := agentgo.GenerateText(params)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, output.ModelName, modelName, "should have correct model name")
	assert.Equal(t, output.Text, result, "should have correct output")
	assert.Equal(t, output.Usage.InputTokens, 245, "should have correct input tokens")
	assert.Equal(t, output.Usage.OutputTokens, 123, "should have correct output tokens")
	assert.Equal(t, output.Usage.CachedTokens, 21, "should have correct cached token")
	assert.Equal(t, output.Usage.ReasoningTokens, 12, "should have correct reasoning token")
}

func TestGenerateTextWithTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Say this is a test"
	modelName := "mocked-llm-3.6-flash"
	result := "This is a test"

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := agentgo.Params{
		Prompt:   prompt,
		Provider: mockProvider,
		EndConditions: []agentgo.EndCondition{
			agentgo.NewMaxStepsEndCondition(5),
		},
		Tools: []models.Tool{
			models.NewTool(models.NewToolParams{
				Name: "mock_tool",
				Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
					log.Printf("Tool called with params: %v", params)
					return models.ToolExecuteOutput{}
				},
			}),
		},
	}

	mockResolveToolCallReturn := func(returned *models.ToolCall) *gomock.Call {
		return mockProvider.EXPECT().ResolveToolCall(gomock.Any(), gomock.Any()).Return(returned)
	}
	gomock.InOrder(
		mockResolveToolCallReturn(&models.ToolCall{
			ToolName: "mock_tool",
			Params: map[string]any{
				"param1": "value1",
				"param2": "value2",
			},
		}),
		mockResolveToolCallReturn(nil),
	)
	mockProvider.EXPECT().Context().MinTimes(1).Return(models.LanguageModelContext{
		ModelName: modelName,
	})
	mockProvider.EXPECT().GenerateText(gomock.Cond(
		func(p providers.AgentProviderPromptMessageParams) bool {
			if len(p.Messages) != 2 {
				return false
			}

			if p.Messages[0].Content().Text() != params.Prompt {
				return false
			}

			return true
		}),
	).Return(
		models.LanguageModelOutput{
			Text: result,
			Usage: models.LanguageModelUsage{
				OutputTokens:    123,
				InputTokens:     245,
				CachedTokens:    21,
				ReasoningTokens: 12,
			},
			ModelName: modelName,
		},
		nil,
	)
	output, err := agentgo.GenerateText(params)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, modelName, output.ModelName, "should have correct model name")
	assert.Equal(t, result, output.Text, "should have correct output")
	assert.Equal(t, output.Usage.InputTokens, 245, "should have correct input tokens")
	assert.Equal(t, output.Usage.OutputTokens, 123, "should have correct output tokens")
	assert.Equal(t, output.Usage.CachedTokens, 21, "should have correct cached token")
	assert.Equal(t, output.Usage.ReasoningTokens, 12, "should have correct reasoning token")
}
