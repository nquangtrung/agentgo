package agentgo

import (
	"context"
	"fmt"
	"testing"

	"github.com/nquangtrung/agentgo/endconditions"
	"github.com/nquangtrung/agentgo/mocks"
	"github.com/nquangtrung/agentgo/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// ============================================================================
// RETRY AND ERROR HANDLING TESTS
// ============================================================================

func TestGenerateTextToolNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Call a tool that doesn't exist"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "existing_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"result": "ok"},
					Usage: models.LanguageModelUsage{
						OutputTokens: 10,
						InputTokens:  20,
						TotalTokens:  30,
					},
				}
			},
		}),
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	// First call returns a tool call for a non-existent tool
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Eq(tools),
	).Return(models.LanguageModelToolCallResolveOutput{
		ToolCalls: []models.ToolCall{{ToolName: "non_existent_tool", Params: map[string]any{}}},
	}, nil).Times(1)
	// Second call (after retry) returns no tool calls
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Eq(tools),
	).Return(models.LanguageModelToolCallResolveOutput{}, nil).Times(1)
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelOutput{Text: "done", ModelName: modelName}, nil)

	output, err := GenerateText(ctx, Params{
		Prompt:        prompt,
		Provider:      mockProvider,
		EndConditions: []models.EndCondition{endconditions.NewMaxStepsEndCondition(10)},
		Tools:         tools,
	})

	assert.NoError(t, err)
	assert.Equal(t, modelName, output.ModelName, "should have correct model name")
	// Tool not found error should be recorded but execution continues
	assert.True(t, len(output.Context.Records()) > 0, "should have execution records")
}

func TestGenerateTextToolExecutionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Call a tool that fails"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	toolExecutionError := fmt.Errorf("tool crashed")
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "failing_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: nil,
					Error:  toolExecutionError,
					Usage: models.LanguageModelUsage{
						OutputTokens: 10,
						InputTokens:  20,
						TotalTokens:  30,
					},
				}
			},
		}),
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	gomock.InOrder(
		mockProvider.EXPECT().ResolveToolCall(
			gomock.Any(),
			gomock.Any(),
			gomock.Eq(tools),
		).Return(models.LanguageModelToolCallResolveOutput{
			ToolCalls: []models.ToolCall{{ToolName: "failing_tool", Params: map[string]any{}}},
		}, nil),
		mockProvider.EXPECT().ResolveToolCall(
			gomock.Any(),
			gomock.Any(),
			gomock.Eq(tools),
		).Return(models.LanguageModelToolCallResolveOutput{}, nil),
	)
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelOutput{Text: "handled error", ModelName: modelName}, nil)

	output, err := GenerateText(ctx, Params{
		Prompt:        prompt,
		Provider:      mockProvider,
		EndConditions: []models.EndCondition{endconditions.NewMaxStepsEndCondition(10)},
		Tools:         tools,
	})

	assert.NoError(t, err)
	assert.Equal(t, "handled error", output.Text, "should continue after tool error")
	assert.Equal(t, modelName, output.ModelName)
}

func TestGenerateTextResolveToolCallError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Provider fails to resolve tools"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "test_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"ok": true},
					Usage:  models.LanguageModelUsage{OutputTokens: 10, InputTokens: 20, TotalTokens: 30},
				}
			},
		}),
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	// First call fails (triggers retry)
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Eq(tools),
	).Return(models.LanguageModelToolCallResolveOutput{}, fmt.Errorf("provider error")).Times(1)
	// After retry, returns successfully
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Eq(tools),
	).Return(models.LanguageModelToolCallResolveOutput{}, nil).Times(1)
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelOutput{Text: "recovered", ModelName: modelName}, nil)

	output, err := GenerateText(ctx, Params{
		Prompt:        prompt,
		Provider:      mockProvider,
		EndConditions: []models.EndCondition{endconditions.NewMaxStepsEndCondition(10)},
		Tools:         tools,
	})

	assert.NoError(t, err)
	assert.Equal(t, "recovered", output.Text, "should recover from provider error")
	assert.Equal(t, modelName, output.ModelName)
}
