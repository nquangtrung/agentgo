package deepseek

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/providers/deepseek/mocks"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewDeepSeekProvider(t *testing.T) {
	provider := NewDeepSeekProvider("test-key", "deepseek-flash")
	assert.Equal(t, "deepseek-flash", provider.Context().ModelName)
	assert.Equal(t, "https://api.deepseek.com", defaultBaseURL)
}

func TestGenerateText(t *testing.T) {
	mockService := mocks.NewMockdeepSeekResponsesService(gomock.NewController(t))
	provider := newDeepSeekProviderWithClient("deepseek-flash", mockService)
	ctx := context.Background()

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

	output, err := provider.GenerateText(
		ctx,
		providers.AgentProviderPromptMessageParams{
			Messages: []models.Message{
				models.NewSystemStringMessage("You are a helpful assistant."),
				models.NewHumanStringMessage("Hello, how are you?"),
			},
		})

	require.NoError(t, err)
	assert.Equal(t, "Great, thank you!", output.Text)
	assert.Equal(t, "deepseek-flash", output.ModelName)
	assert.Equal(t, models.LanguageModelUsage{
		InputTokens: 125,
		InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
			CachedTokens:     30,
			CacheWriteTokens: 20,
		},
		OutputTokens: 256,
		OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
			ReasoningTokens: 50,
		},
		TotalTokens: 381,
	}, output.Usage)
}

func TestGenerateTextTotalTokensFallback(t *testing.T) {
	mockService := mocks.NewMockdeepSeekResponsesService(gomock.NewController(t))
	provider := newDeepSeekProviderWithClient("deepseek-flash", mockService)

	// DeepSeek omits total_tokens on some responses, so the provider derives it.
	mockService.EXPECT().New(gomock.Any(), gomock.Any()).Return(&responses.Response{
		Output: []responses.ResponseOutputItemUnion{{
			Content: []responses.ResponseOutputMessageContentUnion{{Text: "hi", Type: "output_text"}},
		}},
		Usage: responses.ResponseUsage{InputTokens: 10, OutputTokens: 5},
	}, nil)

	output, err := provider.GenerateText(context.Background(), providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{models.NewHumanStringMessage("hi")},
	})

	require.NoError(t, err)
	assert.Equal(t, int64(15), output.Usage.TotalTokens)
}

func TestGenerateTextPropagatesError(t *testing.T) {
	mockService := mocks.NewMockdeepSeekResponsesService(gomock.NewController(t))
	provider := newDeepSeekProviderWithClient("deepseek-flash", mockService)

	mockService.EXPECT().New(gomock.Any(), gomock.Any()).Return(nil, io.ErrUnexpectedEOF)

	_, err := provider.GenerateText(context.Background(), providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{models.NewHumanStringMessage("hi")},
	})

	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func TestResolveToolCall(t *testing.T) {
	mockService := mocks.NewMockdeepSeekResponsesService(gomock.NewController(t))
	provider := newDeepSeekProviderWithClient("deepseek-flash", mockService)
	ctx := context.Background()

	mockArgsStr := `{"key1":"value1","key2":"value2"}`
	mockService.EXPECT().New(gomock.Any(), gomock.Any()).Return(&responses.Response{
		ID: "mocked-response-id",
		Output: []responses.ResponseOutputItemUnion{{
			Type: "function_call", Name: "mock-tool-1",
			Arguments: responses.ResponseOutputItemUnionArguments{OfString: mockArgsStr},
		}},
		Usage: responses.ResponseUsage{InputTokens: 10, OutputTokens: 4},
	}, nil)

	mockTool := models.NewTool(models.NewToolParams{
		Name: "mock-tool-1",
		Fn: func(p models.ToolExecuteParams) models.ToolExecuteOutput {
			return models.ToolExecuteOutput{Output: map[string]any{"key1": "result1"}}
		},
	})

	output, err := provider.ResolveToolCall(
		ctx,
		providers.AgentProviderPromptMessageParams{Messages: []models.Message{models.NewHumanStringMessage("Hello")}},
		[]models.BaseTool{mockTool},
	)

	require.NoError(t, err)
	require.Len(t, output.ToolCalls, 1)
	assert.Equal(t, "mock-tool-1", output.ToolCalls[0].ToolName)
	assert.Equal(t, map[string]any{"key1": "value1", "key2": "value2"}, output.ToolCalls[0].Params)
	assert.Equal(t, int64(10), output.Usage.InputTokens)
	assert.Equal(t, "deepseek-flash", output.ModelName)
}

// TestStreamText drives StreamText with a real SSE decoder instead of a mock,
// which verifies DeepSeek's termination behaviour: the stream ends with a
// `response.completed` event and no `[DONE]` sentinel.
func TestStreamText(t *testing.T) {
	sse := strings.Join([]string{
		`event: response.created`,
		`data: {"type":"response.created"}`,
		``,
		`event: response.output_text.delta`,
		`data: {"type":"response.output_text.delta","delta":"Deep"}`,
		``,
		`event: response.output_text.delta`,
		`data: {"type":"response.output_text.delta","delta":"Seek"}`,
		``,
		`event: response.completed`,
		`data: {"type":"response.completed","response":{"usage":{"input_tokens":12,"output_tokens":3,"input_tokens_details":{"cached_tokens":8},"output_tokens_details":{"reasoning_tokens":2}}}}`,
		``,
		``,
	}, "\n")

	res := &http.Response{
		Header: http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:   io.NopCloser(strings.NewReader(sse)),
	}

	mockService := mocks.NewMockdeepSeekResponsesService(gomock.NewController(t))
	mockService.EXPECT().NewStreaming(gomock.Any(), gomock.Any()).Return(
		ssestream.NewStream[responses.ResponseStreamEventUnion](ssestream.NewDecoder(res), nil),
	)

	provider := newDeepSeekProviderWithClient("deepseek-flash", mockService)
	channel := make(chan models.Part, 8)
	emitter := models.NewPartEmitter(channel)

	output, err := provider.StreamText(context.Background(), providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{models.NewHumanStringMessage("hi")},
	}, *emitter)

	require.NoError(t, err)
	close(channel)

	assert.Equal(t, "DeepSeek", output.Text)
	assert.Equal(t, "deepseek-flash", output.ModelName)
	assert.Equal(t, int64(12), output.Usage.InputTokens)
	assert.Equal(t, int64(3), output.Usage.OutputTokens)
	assert.Equal(t, int64(15), output.Usage.TotalTokens)
	assert.Equal(t, int64(8), output.Usage.InputTokensDetails.CachedTokens)
	assert.Equal(t, int64(2), output.Usage.OutputTokensDetails.ReasoningTokens)

	var emitted []models.Part
	for part := range channel {
		emitted = append(emitted, part)
	}
	require.Len(t, emitted, 4)

	_, ok := emitted[0].(models.StepStartPart)
	assert.True(t, ok, "first part should be a step start")

	text, ok := emitted[1].(models.TextPart)
	require.True(t, ok, "second part should be text")
	assert.Equal(t, "Deep", text.Text())

	text, ok = emitted[2].(models.TextPart)
	require.True(t, ok, "third part should be text")
	assert.Equal(t, "Seek", text.Text())

	_, ok = emitted[3].(models.StepEndPart)
	assert.True(t, ok, "last part should be a step end")
}
