package deepseek

import (
	"testing"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/openai/openai-go/v3/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// toolCallMessage builds the assistant turn a provider would have produced.
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

func TestConvertMessageObjectToInputToolCall(t *testing.T) {
	input := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{
			models.NewHumanStringMessage("weather in Bonn?"),
			toolCallMessage(),
		},
	})

	items := input.OfInputItemList
	require.Len(t, items, 2)

	// The human turn stays a message item.
	assert.Equal(t, responses.EasyInputMessageRoleUser, items[0].OfMessage.Role)

	// The model's own request replays as a function_call item so the API can
	// match it against the result that follows.
	functionCall := items[1].OfFunctionCall
	require.NotNil(t, functionCall, "tool call should become a function_call item")
	assert.Equal(t, "call_1", functionCall.CallID)
	assert.Equal(t, "get_weather", functionCall.Name)
	assert.JSONEq(t, `{"city":"Bonn"}`, functionCall.Arguments)
}

func TestConvertMessageObjectToInputToolResult(t *testing.T) {
	input := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{
			models.NewHumanStringMessage("weather in Bonn?"),
			toolCallMessage(),
			toolResultMessage(),
		},
	})

	items := input.OfInputItemList
	require.Len(t, items, 3)

	output := items[2].OfFunctionCallOutput
	require.NotNil(t, output, "tool result should become a function_call_output item")
	assert.Equal(t, "call_1", output.CallID)
	assert.JSONEq(t, `{"temp":"23C"}`, output.Output.OfString.Value)
}

func TestConvertMessageObjectToInputParallelToolResults(t *testing.T) {
	firstResult := models.NewMessageFromToolResult(models.ToolExecuteOutput{
		Output:   map[string]any{"temp": "23C"},
		ToolCall: &models.ToolCall{ToolName: "get_weather", ID: "call_1"},
	})
	secondResult := models.NewMessageFromToolResult(models.ToolExecuteOutput{
		Output:   map[string]any{"temp": "19C"},
		ToolCall: &models.ToolCall{ToolName: "get_weather", ID: "call_2"},
	})

	input := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{
			models.NewAssistantToolCallsMessage([]models.ToolCall{
				{ToolName: "get_weather", ID: "call_1", Params: map[string]any{"city": "Bonn"}},
				{ToolName: "get_weather", ID: "call_2", Params: map[string]any{"city": "Berlin"}},
			}, ""),
			firstResult,
			secondResult,
		},
	})

	items := input.OfInputItemList
	require.Len(t, items, 4)

	assert.Equal(t, "call_1", items[0].OfFunctionCall.CallID)
	assert.Equal(t, "call_2", items[1].OfFunctionCall.CallID)
	assert.Equal(t, "call_1", items[2].OfFunctionCallOutput.CallID)
	assert.Equal(t, "call_2", items[3].OfFunctionCallOutput.CallID)
}

func TestConvertMessageObjectToInputToolCallWithNarration(t *testing.T) {
	withText := models.NewAssistantToolCallsMessage(
		[]models.ToolCall{{ToolName: "get_weather", ID: "call_1", Params: map[string]any{}}},
		"Let me check.",
	)

	input := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{withText},
	})

	items := input.OfInputItemList
	require.Len(t, items, 2, "narration becomes its own output message item")

	require.NotNil(t, items[0].OfFunctionCall)
	require.NotNil(t, items[1].OfOutputMessage)
	assert.Equal(t, "Let me check.", items[1].OfOutputMessage.Content[0].OfOutputText.Text)
}

func TestConvertMessageObjectToInputToolResultError(t *testing.T) {
	failed := models.NewMessageFromToolResult(models.ToolExecuteOutput{
		Error:    assert.AnError,
		ToolCall: &models.ToolCall{ToolName: "get_weather", ID: "call_1"},
	})

	input := convertInputFromParams(providers.AgentProviderPromptMessageParams{
		Messages: []models.Message{failed},
	})

	items := input.OfInputItemList
	require.Len(t, items, 1)
	require.NotNil(t, items[0].OfFunctionCallOutput)
	assert.Contains(t, items[0].OfFunctionCallOutput.Output.OfString.Value, assert.AnError.Error())
}
