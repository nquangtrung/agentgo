package agentgo

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/nquangtrung/agentgo/endconditions"
	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/mocks"
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestToolMiddlewareOutputValidationFailed verifies that when a ToolMiddleware.After hook
// returns an error (output validation failure), the tool still executed, the error is recorded,
// and the pipeline continues.
func TestToolMiddlewareOutputValidationFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Call mock_tool"
	modelName := "mocked-llm-3.6-flash"
	resultText := "done"
	ctx := context.Background()

	// Capture whether hooks and tool were called
	var toolFnCalled bool
	var afterHookCalled bool

	toolParams := map[string]any{"param": "value"}
	toolResult := map[string]any{"score": 0.3}

	// Create mock tool
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "mock_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				toolFnCalled = true
				return models.ToolExecuteOutput{
					Output: toolResult,
					Usage: models.LanguageModelUsage{
						OutputTokens: 50,
						InputTokens:  80,
					},
				}
			},
		}),
	}

	// Create middleware that validates output
	toolMiddlewares := []ToolMiddleware{
		{
			Before: nil, // No Before hook for this test
			After: func(ctx context.Context, mwCtx ToolMiddlewareContext) error {
				afterHookCalled = true
				// Check output score
				if score, ok := mwCtx.Result.Output["score"].(float64); ok && score < 0.7 {
					return fmt.Errorf("validation failed: score too low (got %.2f, need >= 0.7)", score)
				}
				return nil
			},
		},
	}

	// Mock provider
	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := Params{
		Prompt:          prompt,
		Provider:        mockProvider,
		Tools:           tools,
		ToolMiddlewares: toolMiddlewares,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})

	// First ResolveToolCall returns tool
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
			return len(p.Messages) == 1 && p.Messages[0].Content().Text() == prompt
		}),
		gomock.Eq(tools),
	).Return(
		models.LanguageModelToolCallResolveOutput{
			ToolCalls: []models.ToolCall{
				{ToolName: "mock_tool", Params: toolParams},
			},
		},
		nil,
	).Times(1)

	// Second ResolveToolCall returns empty (no more tools)
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Eq(tools),
	).Return(
		models.LanguageModelToolCallResolveOutput{},
		nil,
	).Times(1)

	// GenerateText call
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(
		models.LanguageModelOutput{
			Text:      resultText,
			ModelName: modelName,
			Usage: models.LanguageModelUsage{
				OutputTokens: 100,
				InputTokens:  150,
			},
		},
		nil,
	).Times(1)

	// Execute
	output, err := GenerateText(ctx, params)

	// Assertions
	assert.NoError(t, err, "should not return error despite middleware validation failure")
	assert.Equal(t, resultText, output.Text, "should have correct final text")
	assert.True(t, toolFnCalled, "tool function should have been called")
	assert.True(t, afterHookCalled, "After hook should have been called")
	assert.Len(t, output.Context.Records(), 2, "should have 2 records (tool + text)")

	// Verify tool record has the validation error
	toolRecord := output.Context.Records()[0]
	assert.NotNil(t, toolRecord.ToolResult.Error, "tool result should have error")
	assert.ErrorContains(t, toolRecord.ToolResult.Error, "validation failed", "error should mention validation")
}

// TestToolMiddlewareInputValidationFailed verifies that when a ToolMiddleware.Before hook
// returns an error (input validation failure), the tool never executes, the error is recorded,
// the After hook is not called, and the pipeline continues.
func TestToolMiddlewareInputValidationFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Call mock_tool"
	modelName := "mocked-llm-3.6-flash"
	resultText := "done"
	ctx := context.Background()

	// Capture whether hooks and tool were called
	var toolFnCallCount int
	var beforeHookCallCount int
	var afterHookCallCount int

	toolParams := map[string]any{} // Missing required param "x"

	// Create mock tool
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "mock_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				toolFnCallCount++
				return models.ToolExecuteOutput{
					Output: map[string]any{"status": "ok"},
					Usage: models.LanguageModelUsage{
						OutputTokens: 50,
						InputTokens:  80,
					},
				}
			},
		}),
	}

	// Create middleware that validates input
	toolMiddlewares := []ToolMiddleware{
		{
			Before: func(ctx context.Context, mwCtx ToolMiddlewareContext) error {
				beforeHookCallCount++
				// Check for required param "x"
				if _, ok := mwCtx.Params["x"]; !ok {
					return fmt.Errorf("input validation failed: param 'x' is required")
				}
				return nil
			},
			After: func(ctx context.Context, mwCtx ToolMiddlewareContext) error {
				afterHookCallCount++
				return nil
			},
		},
	}

	// Mock provider
	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := Params{
		Prompt:          prompt,
		Provider:        mockProvider,
		Tools:           tools,
		ToolMiddlewares: toolMiddlewares,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})

	// First ResolveToolCall returns tool with missing param
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
			return len(p.Messages) == 1 && p.Messages[0].Content().Text() == prompt
		}),
		gomock.Eq(tools),
	).Return(
		models.LanguageModelToolCallResolveOutput{
			ToolCalls: []models.ToolCall{
				{ToolName: "mock_tool", Params: toolParams},
			},
		},
		nil,
	).Times(1)

	// Second ResolveToolCall returns empty
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Eq(tools),
	).Return(
		models.LanguageModelToolCallResolveOutput{},
		nil,
	).Times(1)

	// GenerateText call
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(
		models.LanguageModelOutput{
			Text:      resultText,
			ModelName: modelName,
			Usage: models.LanguageModelUsage{
				OutputTokens: 100,
				InputTokens:  150,
			},
		},
		nil,
	).Times(1)

	// Execute
	output, err := GenerateText(ctx, params)

	// Assertions
	assert.NoError(t, err, "should not return error despite middleware validation failure")
	assert.Equal(t, resultText, output.Text, "should have correct final text")
	assert.Equal(t, 0, toolFnCallCount, "tool function should never be called")
	assert.Equal(t, 1, beforeHookCallCount, "Before hook should be called exactly once")
	assert.Equal(t, 0, afterHookCallCount, "After hook should never be called (tool didn't execute)")
	assert.Len(t, output.Context.Records(), 2, "should have 2 records (error + text)")

	// Verify tool record has the validation error
	toolRecord := output.Context.Records()[0]
	assert.NotNil(t, toolRecord.ToolResult.Error, "tool result should have error")
	assert.ErrorContains(t, toolRecord.ToolResult.Error, "input validation failed", "error should mention input validation")
}

// TestToolMiddlewareHumanInTheLoopInterrupt verifies that when a ToolMiddleware.Before hook
// calls graph.Interrupt, the execution pauses and returns an interrupt error with the correct
// context for user approval workflow.
func TestToolMiddlewareHumanInTheLoopInterrupt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Call dangerous_tool"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	// Capture whether hook was called
	var beforeHookCallCount int

	toolParams := map[string]any{"action": "delete"}

	// Create mock tool
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "dangerous_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				// Should not be called - interrupt happens before tool execution
				panic("tool should not be called")
			},
		}),
	}

	// Create middleware that requires human approval
	toolMiddlewares := []ToolMiddleware{
		{
			Before: func(ctx context.Context, mwCtx ToolMiddlewareContext) error {
				beforeHookCallCount++
				_, err := graph.Interrupt[agentState, agentStateDelta](
					ctx,
					"approve-dangerous-tool",
					map[string]any{
						"tool":   mwCtx.ToolName,
						"params": mwCtx.Params,
					},
				)
				return err
			},
		},
	}

	// Mock provider
	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := Params{
		Prompt:          prompt,
		Provider:        mockProvider,
		Tools:           tools,
		ToolMiddlewares: toolMiddlewares,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})

	// ResolveToolCall returns tool
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
			return len(p.Messages) == 1 && p.Messages[0].Content().Text() == prompt
		}),
		gomock.Eq(tools),
	).Return(
		models.LanguageModelToolCallResolveOutput{
			ToolCalls: []models.ToolCall{
				{ToolName: "dangerous_tool", Params: toolParams},
			},
		},
		nil,
	).Times(1)

	// ===== FIRST INVOCATION =====
	_, err := GenerateText(ctx, params)

	// Should have interrupt error
	assert.Error(t, err, "invocation should error with interrupt")
	superStepErr := new(graph.SuperStepExecutionError)
	assert.True(t, errors.As(err, &superStepErr), "error should be SuperStepExecutionError")

	interrupts := superStepErr.Interrupts()
	assert.Len(t, interrupts, 1, "should have exactly 1 interrupt")
	assert.Equal(t, "approve-dangerous-tool", interrupts[0].Name, "interrupt name should match")
	assert.NotEmpty(t, interrupts[0].ThreadID, "interrupt should have threadID set")

	// Verify middleware was called
	assert.Equal(t, 1, beforeHookCallCount, "Before hook should have been called once")
}
