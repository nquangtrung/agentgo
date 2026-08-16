package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"trontria.com/agentgo/models"
)

func TestPartStart(t *testing.T) {
	context := models.LanguageModelContext{
		ModelName: "mocked-llm",
	}
	var part models.Part = models.NewStepStartPart(
		context,
		"step start",
	)

	p, ok := part.(models.StepStartPart)
	assert.True(t, ok, "should be able to convert to step start")
	assert.NotNil(t, p, "should be able to access converted step")
	assert.Equal(t, p.StepName(), "step start", "should be correct step name")

	_, ok2 := part.(models.StepEndPart)
	assert.False(t, ok2, "should not be able to convert to other part")
}

func TestTextPart(t *testing.T) {
	context := models.LanguageModelContext{
		ModelName: "mocked-llm",
	}
	var part models.Part = models.NewTextPart(context, "hello world")

	// Test AsTextPart conversion
	p, ok := part.(models.TextPart)
	assert.True(t, ok, "should be able to convert to text part")
	assert.NotNil(t, p, "should be able to access converted text part")
	assert.Equal(t, p.Text(), "hello world", "should have correct text")
	assert.Equal(t, p.Type(), models.PartTypeText, "should have correct part type")

	// Test that other conversions fail
	_, ok = part.(models.StepStartPart)
	assert.False(t, ok, "should not be able to convert text part to step start")
}

func TestToolStartPart(t *testing.T) {
	context := models.LanguageModelContext{
		ModelName: "mocked-llm",
	}
	var part models.Part = models.NewToolStartPart(context, "search_tool")

	// Test AsToolStartPart conversion
	p, ok := part.(models.ToolStartPart)
	assert.True(t, ok, "should be able to convert to tool start part")
	assert.NotNil(t, p, "should be able to access converted tool start part")
	assert.Equal(t, p.ToolName(), "search_tool", "should have correct tool name")
	assert.Equal(t, part.Type(), models.PartTypeToolStart, "should have correct part type")

	// Test that other conversions fail
	_, ok = part.(models.TextPart)
	assert.False(t, ok, "should not be able to convert tool start part to text")
}

func TestToolResultPart(t *testing.T) {
	context := models.LanguageModelContext{
		ModelName: "mocked-llm",
	}
	usage := models.LanguageModelUsage{
		InputTokens:  10,
		OutputTokens: 20,
	}
	result := map[string]any{"result": "success", "data": 42}

	var part models.Part = models.NewToolResultPart(context, "search_tool", result, usage)

	// Test AsToolResultPart conversion
	p, ok := part.(models.ToolResultPart)
	assert.True(t, ok, "should be able to convert to tool result part")
	assert.NotNil(t, p, "should be able to access converted tool result part")
	assert.Equal(t, p.ToolName(), "search_tool", "should have correct tool name")
	assert.Equal(t, p.Type(), models.PartTypeToolResult, "should have correct part type")
	assert.Equal(t, p.Result(), result, "should have correct result data")
	assert.Equal(t, p.Usage(), usage, "should have correct usage")
	assert.Equal(t, p.FinishReason(), models.FinishReasonCompleted, "should have completed finish reason")

	// Test that other conversions fail
	_, ok = part.(models.ToolErrorPart)
	assert.False(t, ok, "should not be able to convert tool result part to tool error")
}

func TestToolErrorPart(t *testing.T) {
	context := models.LanguageModelContext{
		ModelName: "mocked-llm",
	}
	errorData := map[string]any{"error": "tool failed", "code": "ERR_001"}

	var part models.Part = models.NewToolErrorPart(context, "search_tool", errorData)

	// Test AsToolErrorPart conversion
	p, ok := part.(models.ToolErrorPart)
	assert.True(t, ok, "should be able to convert to tool error part")
	assert.NotNil(t, p, "should be able to access converted tool error part")
	assert.Equal(t, p.ToolName(), "search_tool", "should have correct tool name")
	assert.Equal(t, p.Type(), models.PartTypeToolError, "should have correct part type")
	assert.Equal(t, p.Error(), errorData, "should have correct error data")
	assert.Equal(t, p.FinishReason(), models.FinishReasonFailed, "should have failed finish reason")

	// Test that other conversions fail
	_, ok = part.(models.ToolResultPart)
	assert.False(t, ok, "should not be able to convert tool error part to tool result")
}

func TestStepEndPart(t *testing.T) {
	context := models.LanguageModelContext{
		ModelName: "mocked-llm",
	}
	usage := models.LanguageModelUsage{
		InputTokens:  5,
		OutputTokens: 15,
	}
	var part models.Part = models.NewStepEndPart(context, "step end", usage)

	// Test AsStepEndPart conversion
	p, ok := part.(models.StepEndPart)
	assert.True(t, ok, "should be able to convert to step end part")
	assert.NotNil(t, p, "should be able to access converted step end part")
	assert.Equal(t, p.StepName(), "step end", "should have correct step name")
	assert.Equal(t, p.Type(), models.PartTypeStepEnd, "should have correct part type")
	assert.Equal(t, p.Usage(), usage, "should have correct usage")
	assert.Equal(t, p.FinishReason(), models.FinishReasonCompleted, "should have completed finish reason")

	// Test that other conversions fail
	_, ok = part.(models.StepStartPart)
	assert.False(t, ok, "should not be able to convert step end part to step start")
}
