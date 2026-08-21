package openai

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/openai/openai-go/v3/responses"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/providers/openai/mocks"
	"trontria.com/agentgo/utils"
)

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

func TestResolveToolCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockResponsesService := mocks.NewMockopenAIResponsesService(ctrl)
	provider := newOpenAIProviderWithClient(
		"mock-gpt-5.5",
		mockResponsesService,
	)

	mockFunctionCallArguments := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
	mockFunctionCallArgumentsStr := utils.Must(json.Marshal(mockFunctionCallArguments))
	log.Printf("mockFunctionCallArgumentsStr: %s", string(mockFunctionCallArgumentsStr))

	mockResponsesService.EXPECT().
		New(gomock.Any(), gomock.Any()).
		Return(
			&responses.Response{
				ID: "mocked-response-id",
				Output: []responses.ResponseOutputItemUnion{
					{
						Type: "function_call",
						Arguments: responses.ResponseOutputItemUnionArguments{
							OfString: string(mockFunctionCallArgumentsStr),
						},
						Name: "mock-tool-1",
					},
				},
			},
			nil,
		)

	mockTool1 := models.NewTool(models.NewToolParams{
		Name:         "mock-tool-1",
		Description:  "This is a description for mock-tool-1",
		InputSchema:  map[string]any{},
		OutputSchema: map[string]any{},
		Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
			return models.ToolExecuteOutput{
				Result: map[string]any{
					"key1": "result1",
					"key2": "result2",
				},
				Usage: models.LanguageModelUsage{
					OutputTokens:    123,
					InputTokens:     134,
					CachedTokens:    145,
					ReasoningTokens: 156,
				},
			}
		},
	})
	toolCall, err := provider.ResolveToolCall(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{
			models.NewSystemStringMessage("You are a helpful assistant."),
			models.NewHumanStringMessage("Hello, how are you?"),
		},
	}, []models.BaseTool{
		mockTool1,
	})

	assert.Nil(t, err, "expect no error is returned")
	assert.Len(t, toolCall, 1, "expect tool call length is correct")
	assert.Equal(t, "mock-tool-1", toolCall[0].ToolName, "expect tool name is correct")
	assert.Equal(t, mockFunctionCallArguments, toolCall[0].Params, "expect tool params are correct")
}
