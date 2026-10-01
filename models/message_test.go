package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStringMessage(t *testing.T) {
	message := NewHumanStringMessage("hello")

	assert.Equal(t, MessageRoleHuman, message.Type())
	assert.Equal(t, "hello", message.Content().Text())
	require.Len(t, message.Content().Parts(), 1)
	assert.Equal(t, MessageContentPartText, message.Content().Parts()[0].PartType())
}

func TestContentTextJoinsMultipleTextParts(t *testing.T) {
	message := NewAssistantToolCallsMessage(
		[]ToolCall{{ToolName: "get_weather", ID: "call_1", Params: map[string]any{"city": "Bonn"}}},
		"Let me check that.",
	)

	content := message.Content()
	// The tool call is excluded from Text; only the narration is returned.
	assert.Equal(t, "Let me check that.", content.Text())

	parts := content.Parts()
	require.Len(t, parts, 2)
	assert.Equal(t, MessageContentPartText, parts[0].PartType())
	assert.Equal(t, MessageContentPartToolCall, parts[1].PartType())
}

func TestContentTextEmptyWithoutTextParts(t *testing.T) {
	message := NewAssistantToolCallsMessage(
		[]ToolCall{{ToolName: "get_weather", ID: "call_1"}},
		"",
	)

	assert.Equal(t, "", message.Content().Text())
	assert.False(t, message.Content().IsEmpty())
	assert.Len(t, message.Content().ToolCalls(), 1)
	assert.Empty(t, message.Content().ToolResults())
}

func TestToolCallContentPart(t *testing.T) {
	input := map[string]any{"city": "Bonn"}
	part := NewToolCallContentPart("call_1", "get_weather", input)

	assert.Equal(t, MessageContentPartToolCall, part.PartType())
	assert.Equal(t, "call_1", part.ID())
	assert.Equal(t, "get_weather", part.Name())
	assert.Equal(t, input, part.Input())
}

func TestNewMessageFromToolResult(t *testing.T) {
	output := ToolExecuteOutput{
		Output:   map[string]any{"temp": "23C"},
		ToolCall: &ToolCall{ToolName: "get_weather", ID: "call_1"},
	}

	message := NewMessageFromToolResult(output)

	assert.Equal(t, MessageRoleTool, message.Type())

	results := message.Content().ToolResults()
	require.Len(t, results, 1)
	assert.Equal(t, "call_1", results[0].ToolCallID())
	assert.Equal(t, "get_weather", results[0].Name())
	assert.Equal(t, map[string]any{"temp": "23C"}, results[0].Output())
	assert.NoError(t, results[0].Err())
	assert.JSONEq(t, `{"temp":"23C"}`, results[0].Payload())
}

func TestNewMessageFromToolResultError(t *testing.T) {
	output := ToolExecuteOutput{
		Error:    errors.New("boom"),
		ToolCall: &ToolCall{ToolName: "get_weather", ID: "call_1"},
	}

	message := NewMessageFromToolResult(output)
	results := message.Content().ToolResults()
	require.Len(t, results, 1)

	assert.Error(t, results[0].Err())
	assert.Contains(t, results[0].Payload(), "boom")
}

func TestNewMessageFromToolResultWithoutToolCall(t *testing.T) {
	message := NewMessageFromToolResult(ToolExecuteOutput{Output: map[string]any{"text": "done"}})

	results := message.Content().ToolResults()
	require.Len(t, results, 1)
	assert.Equal(t, "text", results[0].Name())
	assert.Equal(t, "", results[0].ToolCallID())
}

func TestContentIsEmpty(t *testing.T) {
	assert.True(t, NewMessageContent().IsEmpty())
	assert.False(t, NewMessageContent(NewTextContentPart("x")).IsEmpty())
}

func TestMessageRoleToolIsDistinct(t *testing.T) {
	// Guards against a tool result being mistaken for user input.
	assert.NotEqual(t, MessageRoleHuman, MessageRoleTool)
	assert.NotEqual(t, MessageRoleAssistant, MessageRoleTool)
}
