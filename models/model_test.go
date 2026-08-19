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
	assert.Equal(t, len(execCtx.steps), 0, "expected steps to be empty")

	execCtx.AddStep("step1", nil)
	assert.Equal(t, 1, len(execCtx.steps), "expected steps to have 1 step")
	assert.Equal(t, "step1", execCtx.LastStep().Name, "expected step name to be 'step1'")

	execCtx.UpdateLastStepResult(&ToolExecuteOutput{
		Result: map[string]any{"key": "value"},
	})
	assert.Equal(t, "value", execCtx.LastStep().ToolResult.Result["key"], "expected last step result to have key 'value'")

	execCtx.AddStep("step2", &ToolCall{
		ToolName: "test-tool",
		Params: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	})
	assert.Equal(t, 2, len(execCtx.steps), "expected steps to have 2 steps")
	assert.Equal(t, "step2", execCtx.LastStep().Name, "expected step name to be 'step2'")
	assert.Equal(t, "test-tool", execCtx.LastStep().ToolCalled.ToolName, "expected tool name to be 'test-tool'")
	assert.Equal(t, map[string]any{"key1": "value1", "key2": "value2"}, execCtx.LastStep().ToolCalled.Params, "expected tool params to be correct")
	execCtx.UpdateLastStepError(assert.AnError)
	assert.Equal(t, assert.AnError, execCtx.LastStep().ToolResult.Error, "expected last step error to be assert.AnError")
}
