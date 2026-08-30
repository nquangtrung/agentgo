package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExecutionContextFromLanguageModelContext(t *testing.T) {
	ctx := LanguageModelContext{
		ModelName: "test-model",
	}
	execCtx := NewExecutionContextFromLanguageModelContext(ctx)
	assert.Equal(t, "test-model", execCtx.ModelName(), "expected model name to be 'test-model'")
	assert.Equal(t, len(execCtx.records), 0, "expected steps to be empty")

	execCtx.AddToolCall("step1", nil)
	assert.Equal(t, 1, len(execCtx.records), "expected steps to have 1 step")
	assert.Equal(t, "step1", execCtx.LastRecord().Name, "expected step name to be 'step1'")

	execCtx.UpdateLastRecordResult(&ToolExecuteOutput{
		Output: map[string]any{"key": "value"},
	})
	assert.Equal(t, "value", execCtx.LastRecord().ToolResult.Output["key"], "expected last step output to have key 'value'")

	execCtx.AddToolCall("step2", &ToolCall{
		ToolName: "test-tool",
		Params: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	})
	assert.Equal(t, 2, len(execCtx.records), "expected steps to have 2 steps")
	assert.Equal(t, "step2", execCtx.LastRecord().Name, "expected step name to be 'step2'")
	assert.Equal(t, "test-tool", execCtx.LastRecord().ToolCalled.ToolName, "expected tool name to be 'test-tool'")
	assert.Equal(t, map[string]any{"key1": "value1", "key2": "value2"}, execCtx.LastRecord().ToolCalled.Params, "expected tool params to be correct")
	execCtx.UpdateLastRecordError(assert.AnError)
	assert.Equal(t, assert.AnError, execCtx.LastRecord().ToolResult.Error, "expected last step error to be assert.AnError")
}
