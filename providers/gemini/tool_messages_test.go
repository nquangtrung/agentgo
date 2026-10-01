package gemini

import (
	"testing"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func toolCallMessage() models.BaseMessage {
	return models.NewAssistantToolCallsMessage(
		[]models.ToolCall{{
			ToolName: "get_weather",
			ID:       "call_1",
			Params:   map[string]any{"city": "Bonn"},
		}},
		"",
	)
}

func toolResultMessage() models.BaseMessage {
	return models.NewMessageFromToolResult(models.ToolExecuteOutput{
		Output:   map[string]any{"temp": "23C"},
		ToolCall: &models.ToolCall{ToolName: "get_weather", ID: "call_1"},
	})
}

// TestConvertMessageObjectToMessagesToolCall pins the shape Anthropic requires:
// the assistant's request replays with tool_calls, and the result follows as a
// role:"tool" message carrying the matching tool_call_id. The native Messages
// API rejects a tool_result that does not reference a preceding tool_use.
func TestConvertMessageObjectToMessagesToolCall(t *testing.T) {
	converted := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{
			models.NewHumanStringMessage("weather in Bonn?"),
			toolCallMessage(),
		},
	})

	require.Len(t, converted, 2)

	assistant := converted[1].OfAssistant
	require.NotNil(t, assistant, "tool call should ride on an assistant turn")
	require.Len(t, assistant.ToolCalls, 1)
	assert.Equal(t, "call_1", assistant.ToolCalls[0].OfFunction.ID)
	assert.Equal(t, "get_weather", assistant.ToolCalls[0].OfFunction.Function.Name)
	assert.JSONEq(t, `{"city":"Bonn"}`, assistant.ToolCalls[0].OfFunction.Function.Arguments)
}

func TestConvertMessageObjectToMessagesToolResult(t *testing.T) {
	converted := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{
			models.NewHumanStringMessage("weather in Bonn?"),
			toolCallMessage(),
			toolResultMessage(),
		},
	})

	require.Len(t, converted, 3)

	result := converted[2].OfTool
	require.NotNil(t, result, "tool result should become a role:tool message")
	assert.Equal(t, "call_1", result.ToolCallID)
	assert.JSONEq(t, `{"temp":"23C"}`, result.Content.OfString.Value)
}

func TestConvertMessageObjectToMessagesToolResultError(t *testing.T) {
	failed := models.NewMessageFromToolResult(models.ToolExecuteOutput{
		Error:    assert.AnError,
		ToolCall: &models.ToolCall{ToolName: "get_weather", ID: "call_1"},
	})

	converted := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{failed},
	})

	require.Len(t, converted, 1)
	require.NotNil(t, converted[0].OfTool)
	assert.Contains(t, converted[0].OfTool.Content.OfString.Value, assert.AnError.Error())
}

func TestConvertMessageObjectToMessagesParallelToolResults(t *testing.T) {
	converted := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{
			models.NewAssistantToolCallsMessage([]models.ToolCall{
				{ToolName: "get_weather", ID: "call_1", Params: map[string]any{"city": "Bonn"}},
				{ToolName: "get_weather", ID: "toolu_02", Params: map[string]any{"city": "Berlin"}},
			}, ""),
			models.NewMessageFromToolResult(models.ToolExecuteOutput{
				Output:   map[string]any{"temp": "23C"},
				ToolCall: &models.ToolCall{ToolName: "get_weather", ID: "call_1"},
			}),
			models.NewMessageFromToolResult(models.ToolExecuteOutput{
				Output:   map[string]any{"temp": "19C"},
				ToolCall: &models.ToolCall{ToolName: "get_weather", ID: "toolu_02"},
			}),
		},
	})

	require.Len(t, converted, 3)
	require.Len(t, converted[0].OfAssistant.ToolCalls, 2)
	assert.Equal(t, "call_1", converted[1].OfTool.ToolCallID)
	assert.Equal(t, "toolu_02", converted[2].OfTool.ToolCallID)
}

func TestConvertMessageObjectToMessagesToolCallWithNarration(t *testing.T) {
	withText := models.NewAssistantToolCallsMessage(
		[]models.ToolCall{{ToolName: "get_weather", ID: "call_1", Params: map[string]any{}}},
		"Let me check.",
	)

	converted := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{withText},
	})

	require.Len(t, converted, 1)
	assistant := converted[0].OfAssistant
	require.NotNil(t, assistant)
	assert.Equal(t, "Let me check.", assistant.Content.OfString.Value)
	require.Len(t, assistant.ToolCalls, 1)
}

// A call without an id cannot be correlated with its result, so it is dropped
// rather than replayed and leaving a dangling result.
func TestConvertMessageObjectToMessagesSkipsCallsWithoutID(t *testing.T) {
	noID := models.NewAssistantToolCallsMessage(
		[]models.ToolCall{{ToolName: "get_weather", Params: map[string]any{}}},
		"",
	)

	converted := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{noID},
	})

	require.Len(t, converted, 1)
	require.NotNil(t, converted[0].OfAssistant)
	assert.Empty(t, converted[0].OfAssistant.ToolCalls)
}
