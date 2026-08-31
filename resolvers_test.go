package agentgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"
)

func TestResolveExecutionContextAsTextOutput(t *testing.T) {
	ctx := models.LanguageModelContext{
		ModelName: "test-model",
	}
	execCtx := models.NewExecutionContextFromLanguageModelContext(ctx)

	_, err := resolveExecutionContextAsTextOutput(execCtx)
	assert.Error(t, err, "expected error when resolving execution context as text output")

	execCtx.AddToolCall("tool-step: mock_tool_1",
		&models.ToolCall{
			ToolName: "mock_tool_1",
			Params: map[string]any{
				"param1": "value1",
				"param2": "value2",
			},
		},
	)
	execCtx.UpdateLastRecordResult(&models.ToolExecuteOutput{
		Output: map[string]any{
			"key1": "value1",
		},
		Usage: models.LanguageModelUsage{
			OutputTokens: 100,
			InputTokens:  10,
			InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
				CachedTokens: 20,
			},
			TotalTokens: 110,
		},
	})
	execCtx.AddToolCall("tool-step: mock_tool_2",
		&models.ToolCall{
			ToolName: "mock_tool_2",
			Params: map[string]any{
				"param3": "value3",
				"param4": "value4",
			},
		},
	)
	execCtx.UpdateLastRecordResult(&models.ToolExecuteOutput{
		Output: map[string]any{
			"key2": "value2",
		},
		Usage: models.LanguageModelUsage{
			OutputTokens: 200,
			InputTokens:  200,
			InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
				CachedTokens: 20,
			},
			TotalTokens: 400,
		},
	})

	execCtx.AddToolCall("tool-step: mock_tool_3",
		&models.ToolCall{
			ToolName: "mock_tool_3",
			Params: map[string]any{
				"param5": "value5",
				"param6": "value6",
			},
		},
	)
	execCtx.UpdateLastRecordResult(&models.ToolExecuteOutput{
		Output: map[string]any{
			"key3": "value3",
		},
	})

	execCtx.AddToolCall("text-step: final-output", nil)
	execCtx.UpdateLastRecordResult(&models.ToolExecuteOutput{
		Output: map[string]any{
			"text": "final-output",
		},
		Usage: models.LanguageModelUsage{
			OutputTokens: 2200,
			InputTokens:  300,
			InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
				CachedTokens: 20,
			},
			OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
				ReasoningTokens: 20,
			},
			TotalTokens: 2500,
		},
	})

	output, err := resolveExecutionContextAsTextOutput(execCtx)
	assert.NoError(t, err, "expected no error when resolving execution context with steps as text output")
	assert.Equal(t, "test-model", output.ModelName, "expected model name to match")
	assert.Equal(t, "final-output", output.Text, "expected last output to match")
	assert.Equal(t, int64(300+200+10), output.Usage.InputTokens, "expected input tokens to match")
	assert.Equal(t, int64(2200+200+100), output.Usage.OutputTokens, "expected output tokens to match")
	assert.Equal(t, int64(20+20+20), output.Usage.InputTokensDetails.CachedTokens, "expected cached tokens to match")
	assert.Equal(t, int64(20+0+0), output.Usage.OutputTokensDetails.ReasoningTokens, "expected reasoning tokens to match")
	assert.Equal(t, 4, len(output.Context.Records()), "expected number of steps to match")

	assert.Equal(t, "text-step: final-output", utils.Must(output.Context.Record(3)).Name, "expected fourth step name to match")

	assert.Equal(t, "mock_tool_1", utils.Must(output.Context.Record(0)).ToolCalled.ToolName, "expected first step tool name to match")
	assert.Equal(t, "mock_tool_2", utils.Must(output.Context.Record(1)).ToolCalled.ToolName, "expected second step tool name to match")
	assert.Equal(t, "mock_tool_3", utils.Must(output.Context.Record(2)).ToolCalled.ToolName, "expected third step tool name to match")
	assert.Equal(t, "tool-step: mock_tool_1", utils.Must(output.Context.Record(0)).Name, "expected first step name to match")
	assert.Equal(t, "tool-step: mock_tool_2", utils.Must(output.Context.Record(1)).Name, "expected second step name to match")
	assert.Equal(t, "tool-step: mock_tool_3", utils.Must(output.Context.Record(2)).Name, "expected third step name to match")
	assert.Equal(t, map[string]any{"key1": "value1"}, utils.Must(output.Context.Record(0)).ToolResult.Output, "expected first step tool name to match")
	assert.Equal(t, map[string]any{"key2": "value2"}, utils.Must(output.Context.Record(1)).ToolResult.Output, "expected first step tool name to match")
	assert.Equal(t, map[string]any{"key3": "value3"}, utils.Must(output.Context.Record(2)).ToolResult.Output, "expected first step tool name to match")
}
