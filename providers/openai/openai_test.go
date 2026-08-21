package openai

import (
	"testing"

	"github.com/openai/openai-go/v3/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/providers/openai/mocks"
)

func TestGenerateText(t *testing.T) {
	mockService := mocks.NewMockopenAIResponsesService(gomock.NewController(t))
	provider := newOpenAIProviderWithClient("mock-gpt-5.5", mockService)

	mockedResponse := &responses.Response{
		ID: "mocked-response-id",
		Output: []responses.ResponseOutputItemUnion{{
			Content: []responses.ResponseOutputMessageContentUnion{{Text: "Great, thank you!", Type: "output_text"}},
		}},
		Usage: responses.ResponseUsage{
			InputTokens:         125,
			OutputTokens:        256,
			InputTokensDetails:  responses.ResponseUsageInputTokensDetails{CacheWriteTokens: 20, CachedTokens: 30},
			OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{ReasoningTokens: 50},
		},
	}

	mockService.EXPECT().
		New(gomock.Any(), gomock.Any()).
		Return(mockedResponse, nil)

	output, err := provider.GenerateText(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{
			models.NewSystemStringMessage("You are a helpful assistant."),
			models.NewHumanStringMessage("Hello, how are you?"),
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "Great, thank you!", output.Text)
	assert.Equal(t, models.LanguageModelUsage{
		InputTokens:     125,
		OutputTokens:    256,
		CachedTokens:    50, // 20 + 30
		ReasoningTokens: 50,
	}, output.Usage)
}

func TestResolveToolCall(t *testing.T) {
	mockService := mocks.NewMockopenAIResponsesService(gomock.NewController(t))
	provider := newOpenAIProviderWithClient("mock-gpt-5.5", mockService)

	mockArgsStr := `{"key1":"value1","key2":"value2"}`
	mockService.EXPECT().New(gomock.Any(), gomock.Any()).Return(&responses.Response{
		ID: "mocked-response-id",
		Output: []responses.ResponseOutputItemUnion{{
			Type: "function_call", Name: "mock-tool-1",
			Arguments: responses.ResponseOutputItemUnionArguments{OfString: mockArgsStr},
		}},
	}, nil)

	mockTool := models.NewTool(models.NewToolParams{
		Name: "mock-tool-1",
		Fn: func(p models.ToolExecuteParams) models.ToolExecuteOutput {
			return models.ToolExecuteOutput{Result: map[string]any{"key1": "result1"}}
		},
	})

	toolCall, err := provider.ResolveToolCall(
		providers.AgentProviderPromptMessageParams{Messages: []models.Message{models.NewHumanStringMessage("Hello")}},
		[]models.BaseTool{mockTool},
	)

	require.NoError(t, err)
	require.Len(t, toolCall, 1)
	assert.Equal(t, "mock-tool-1", toolCall[0].ToolName)
	assert.Equal(t, map[string]any{"key1": "value1", "key2": "value2"}, toolCall[0].Params)
}
