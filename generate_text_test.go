package agentgo

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"trontria.com/agentgo/mocks"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

func TestGenerateText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Say this is a test"
	modelName := "mocked-llm-3.6-flash"
	result := "This is a test"

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := Params{
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

	output, err := GenerateText(params)
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
	mockProvider := mocks.NewMockAgentProvider(ctrl)

	prompt := "Say this is a test"
	modelName := "mocked-llm-3.6-flash"
	result := "This is a test"

	toolParams := map[string]any{
		"param1": "value1",
		"param2": "value2",
	}
	toolResult := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
	jsonedToolResult := utils.Must(json.Marshal(toolResult))
	tools := []models.Tool{
		models.NewTool(models.NewToolParams{
			Name: "mock_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				log.Printf("Tool called with params: %v", params)
				assert.Equal(t, toolParams, params.Params, "should be called with correct tool params")
				return models.ToolExecuteOutput{
					Result: toolResult,
					Usage: models.LanguageModelUsage{
						OutputTokens:    100,
						InputTokens:     200,
						CachedTokens:    20,
						ReasoningTokens: 0,
					},
				}
			},
		}),
	}
	params := Params{
		Prompt:   prompt,
		Provider: mockProvider,
		EndConditions: []EndCondition{
			NewMaxStepsEndCondition(5),
		},
		Tools: tools,
	}

	checkResolveToolCall := func(p providers.AgentProviderPromptMessageParams) bool {
		if len(p.Messages) != 2 {
			return false
		}

		if p.Messages[0].Content().Text() != prompt {
			return false
		}

		expectedToolCallMessage := fmt.Sprintf("Tool [%s] execution result: %s", tools[0].Name(), jsonedToolResult)
		if p.Messages[1].Content().Text() != expectedToolCallMessage {
			return false
		}

		return true
	}

	mockResolveToolCallReturn := func(
		paramsMatcher gomock.Matcher,
		toolParamsMatcher gomock.Matcher,
		returned *models.ToolCall,
	) *gomock.Call {
		return mockProvider.EXPECT().ResolveToolCall(paramsMatcher, toolParamsMatcher).Return(returned)
	}
	gomock.InOrder(
		mockResolveToolCallReturn(
			gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
				if len(p.Messages) != 1 {
					return false
				}

				return p.Messages[0].Content().Text() == params.Prompt
			}),
			gomock.Eq(tools),
			&models.ToolCall{
				ToolName: "mock_tool",
				Params:   toolParams,
			}),
		mockResolveToolCallReturn(
			gomock.Cond(checkResolveToolCall),
			gomock.Eq(tools),
			nil,
		),
	)
	mockProvider.EXPECT().Context().MinTimes(1).Return(models.LanguageModelContext{
		ModelName: modelName,
	})
	mockProvider.EXPECT().GenerateText(gomock.Cond(checkResolveToolCall)).Return(
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
	output, err := GenerateText(params)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, modelName, output.ModelName, "should have correct model name")
	assert.Equal(t, result, output.Text, "should have correct output")
	assert.Equal(t, 245+200, output.Usage.InputTokens, "should have correct input tokens")
	assert.Equal(t, 123+100, output.Usage.OutputTokens, "should have correct output tokens")
	assert.Equal(t, 21+20, output.Usage.CachedTokens, "should have correct cached token")
	assert.Equal(t, 12+0, output.Usage.ReasoningTokens, "should have correct reasoning token")
	assert.Equal(t, 2, len(output.Context.Steps()), "should have correct number of steps")
	assert.Equal(t, "tool call [mock_tool]", utils.Must(output.Context.Step(0)).Name, "should have correct first step name")
	assert.Equal(t, toolResult, utils.Must(output.Context.Step(0)).ToolResult.Result, "should have correct result in first step")
}
