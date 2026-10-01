package gemini

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/providers/gemini/mocks"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewGeminiProvider(t *testing.T) {
	provider := NewGeminiProvider("test-key", "gemini-3.8-flash")
	assert.Equal(t, "gemini-3.8-flash", provider.Context().ModelName)
	assert.Equal(t, "https://generativelanguage.googleapis.com/v1beta/openai/", defaultBaseURL)
}

func TestGenerateText(t *testing.T) {
	mockService := mocks.NewMockgeminiChatService(gomock.NewController(t))
	provider := newGeminiProviderWithClient("gemini-3.8-flash", mockService)

	mockService.EXPECT().New(gomock.Any(), gomock.Any()).Return(&openai.ChatCompletion{
		ID: "chatcmpl-1",
		Choices: []openai.ChatCompletionChoice{{
			FinishReason: "stop",
			Message: openai.ChatCompletionMessage{
				Content: "Great, thank you!",
				Role:    "assistant",
			},
		}},
		Usage: openai.CompletionUsage{
			PromptTokens:            125,
			CompletionTokens:        256,
			TotalTokens:             381,
			PromptTokensDetails:     openai.CompletionUsagePromptTokensDetails{CachedTokens: 30},
			CompletionTokensDetails: openai.CompletionUsageCompletionTokensDetails{ReasoningTokens: 50},
		},
	}, nil)

	output, err := provider.GenerateText(context.Background(), providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{
			models.NewSystemStringMessage("You are a helpful assistant."),
			models.NewHumanStringMessage("Hello, how are you?"),
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "Great, thank you!", output.Text)
	assert.Equal(t, "gemini-3.8-flash", output.ModelName)
	assert.Equal(t, int64(125), output.Usage.InputTokens)
	assert.Equal(t, int64(256), output.Usage.OutputTokens)
	assert.Equal(t, int64(381), output.Usage.TotalTokens)
	// Gemini reports cache hits and thinking tokens in the usage details.
	assert.Equal(t, int64(30), output.Usage.InputTokensDetails.CachedTokens)
	assert.Equal(t, int64(50), output.Usage.OutputTokensDetails.ReasoningTokens)
}

// TestGenerateTextSendsStructuredOutput pins that Gemini's json_schema support
// is actually wired through, unlike Claude where it is rejected.
func TestGenerateTextSendsStructuredOutput(t *testing.T) {
	mockService := mocks.NewMockgeminiChatService(gomock.NewController(t))
	provider := newGeminiProviderWithClient("gemini-3.8-flash", mockService)

	mockService.EXPECT().New(gomock.Any(), gomock.Any()).Return(&openai.ChatCompletion{
		Choices: []openai.ChatCompletionChoice{{
			FinishReason: "stop",
			Message:      openai.ChatCompletionMessage{Content: `{"name":"pancakes"}`},
		}},
		Usage: openai.CompletionUsage{PromptTokens: 5, CompletionTokens: 8, TotalTokens: 13},
	}, nil)

	output, err := provider.GenerateText(context.Background(), providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{models.NewHumanStringMessage("hi")},
		ResponseFormat: &models.ResponseFormat{
			Name:       "recipe",
			Strict:     true,
			JSONSchema: map[string]any{"type": "object"},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, `{"name":"pancakes"}`, output.Text)
}

func TestGenerateTextPropagatesError(t *testing.T) {
	mockService := mocks.NewMockgeminiChatService(gomock.NewController(t))
	provider := newGeminiProviderWithClient("gemini-3.8-flash", mockService)

	mockService.EXPECT().New(gomock.Any(), gomock.Any()).Return(nil, io.ErrUnexpectedEOF)

	_, err := provider.GenerateText(context.Background(), providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{models.NewHumanStringMessage("hi")},
	})

	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func TestGenerateTextEmptyChoices(t *testing.T) {
	mockService := mocks.NewMockgeminiChatService(gomock.NewController(t))
	provider := newGeminiProviderWithClient("gemini-3.8-flash", mockService)

	mockService.EXPECT().New(gomock.Any(), gomock.Any()).Return(&openai.ChatCompletion{}, nil)

	_, err := provider.GenerateText(context.Background(), providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{models.NewHumanStringMessage("hi")},
	})

	assert.Error(t, err)
}

func TestResolveToolCall(t *testing.T) {
	mockService := mocks.NewMockgeminiChatService(gomock.NewController(t))
	provider := newGeminiProviderWithClient("gemini-3.8-flash", mockService)

	mockService.EXPECT().New(gomock.Any(), gomock.Any()).Return(&openai.ChatCompletion{
		Choices: []openai.ChatCompletionChoice{{
			FinishReason: "tool_calls",
			Message: openai.ChatCompletionMessage{
				Role: "assistant",
				ToolCalls: []openai.ChatCompletionMessageToolCallUnion{{
					ID:   "toolu_01",
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunction{
						Name:      "get-weather",
						Arguments: `{"location":"Bonn"}`,
					},
				}},
			},
		}},
		Usage: openai.CompletionUsage{PromptTokens: 10, CompletionTokens: 4, TotalTokens: 14},
	}, nil)

	tool := models.NewTool(models.NewToolParams{
		Name: "get-weather",
		Fn:   func(p models.ToolExecuteParams) models.ToolExecuteOutput { return models.ToolExecuteOutput{} },
	})

	output, err := provider.ResolveToolCall(
		context.Background(),
		providers.AgentProviderPromptMessageParams{Messages: []models.Message{models.NewHumanStringMessage("weather?")}},
		[]models.BaseTool{tool},
	)

	require.NoError(t, err)
	require.Len(t, output.ToolCalls, 1)
	assert.Equal(t, "get-weather", output.ToolCalls[0].ToolName)
	assert.Equal(t, map[string]any{"location": "Bonn"}, output.ToolCalls[0].Params)
	assert.Equal(t, "toolu_01", output.ToolCalls[0].ID)
	assert.Equal(t, int64(14), output.Usage.TotalTokens)
}

// TestStreamText drives StreamText with a real SSE decoder so the usage-bearing
// final chunk and the empty-delta chunks are exercised end to end.
func TestStreamText(t *testing.T) {
	sse := strings.Join([]string{
		`data: {"id":"1","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":""}]}`,
		``,
		`data: {"id":"1","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":""}]}`,
		``,
		`data: {"id":"1","choices":[{"index":0,"delta":{"content":" there"},"finish_reason":"stop"}]}`,
		``,
		`data: {"id":"1","choices":[],"usage":{"prompt_tokens":9,"completion_tokens":3,"total_tokens":12}}`,
		``,
		`data: [DONE]`,
		``,
		``,
	}, "\n")

	res := &http.Response{
		Header: http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:   io.NopCloser(strings.NewReader(sse)),
	}

	mockService := mocks.NewMockgeminiChatService(gomock.NewController(t))
	mockService.EXPECT().NewStreaming(gomock.Any(), gomock.Any()).Return(
		ssestream.NewStream[openai.ChatCompletionChunk](ssestream.NewDecoder(res), nil),
	)

	provider := newGeminiProviderWithClient("gemini-3.8-flash", mockService)
	channel := make(chan models.Part, 8)
	emitter := models.NewPartEmitter(channel)

	output, err := provider.StreamText(context.Background(), providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{models.NewHumanStringMessage("hi")},
	}, *emitter)

	require.NoError(t, err)
	close(channel)

	assert.Equal(t, "Hello there", output.Text)
	assert.Equal(t, int64(9), output.Usage.InputTokens)
	assert.Equal(t, int64(3), output.Usage.OutputTokens)
	assert.Equal(t, int64(12), output.Usage.TotalTokens)

	var emitted []models.Part
	for part := range channel {
		emitted = append(emitted, part)
	}
	require.Len(t, emitted, 4)

	_, ok := emitted[0].(models.StepStartPart)
	assert.True(t, ok, "first part should be a step start")

	text, ok := emitted[1].(models.TextPart)
	require.True(t, ok)
	assert.Equal(t, "Hello", text.Text())

	text, ok = emitted[2].(models.TextPart)
	require.True(t, ok)
	assert.Equal(t, " there", text.Text())

	_, ok = emitted[3].(models.StepEndPart)
	assert.True(t, ok, "last part should be a step end")
}
