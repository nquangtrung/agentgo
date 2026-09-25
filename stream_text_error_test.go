package agentgo

import (
	"context"
	"fmt"
	"testing"

	"github.com/nquangtrung/agentgo/endconditions"
	"github.com/nquangtrung/agentgo/mocks"
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// ============================================================================
// RETRY AND ERROR HANDLING TESTS
// ============================================================================

func TestStreamTextToolNotFoundError(t *testing.T) {
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
	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, "done"))
	}).Return(models.LanguageModelOutput{Text: "done", ModelName: modelName}, nil)

	output := StreamText(ctx, Params{
		Prompt:        prompt,
		Provider:      mockProvider,
		EndConditions: []models.EndCondition{endconditions.NewMaxStepsEndCondition(10)},
		Tools:         tools,
	})

	var endPart models.ProcessEndPart
	for part := range output.Channel {
		if p, ok := part.(models.ProcessEndPart); ok {
			endPart = p
		}
	}

	assert.Equal(t, modelName, output.ModelName, "should have correct model name")
	assert.NotNil(t, endPart)
}

func TestStreamTextToolExecutionError(t *testing.T) {
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
	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, "handled error"))
	}).Return(models.LanguageModelOutput{Text: "handled error", ModelName: modelName}, nil)

	output := StreamText(ctx, Params{
		Prompt:        prompt,
		Provider:      mockProvider,
		EndConditions: []models.EndCondition{endconditions.NewMaxStepsEndCondition(10)},
		Tools:         tools,
	})

	var endPart models.ProcessEndPart
	for part := range output.Channel {
		if p, ok := part.(models.ProcessEndPart); ok {
			endPart = p
		}
	}

	assert.NotNil(t, endPart)
	assert.Equal(t, modelName, output.ModelName)
}

// ============================================================================
// TABLE-DRIVEN RETRY AND ERROR HANDLING TESTS
// ============================================================================

type streamResolveToolCallErrorTestCase struct {
	name           string
	prompt         string
	errorRetries   int    // number of times to return error before success
	errorType      error  // type of error to return (transient or non-transient)
	shouldSucceed  bool   // whether StreamText should succeed
	expectedOutput string // expected output text if successful
}

func TestStreamTextResolveToolCallError(t *testing.T) {
	modelName := "mocked-llm-3.6-flash"
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

	testCases := []streamResolveToolCallErrorTestCase{
		{
			name:           "transient error once then success",
			prompt:         "Provider fails to resolve tools once",
			errorRetries:   1,
			errorType:      mockTransientError{},
			shouldSucceed:  true,
			expectedOutput: "recovered",
		},
		{
			name:           "transient error twice then success",
			prompt:         "Provider fails to resolve tools twice",
			errorRetries:   2,
			errorType:      mockTransientError{},
			shouldSucceed:  true,
			expectedOutput: "recovered",
		},
		{
			name:           "transient error three times exceeds retry limit",
			prompt:         "Provider fails to resolve tools three times",
			errorRetries:   3,
			errorType:      mockTransientError{},
			shouldSucceed:  false,
			expectedOutput: "",
		},
		{
			name:           "non-transient error fails immediately",
			prompt:         "Provider fails with non-transient error",
			errorRetries:   1,
			errorType:      mockNonTransientError{},
			shouldSucceed:  false,
			expectedOutput: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProvider := mocks.NewMockAgentProvider(ctrl)
			ctx := context.Background()

			mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})

			// Check if this is a transient error (has Timeout() method that returns true)
			isTransient := false
			if transientErr, ok := tc.errorType.(interface{ Timeout() bool }); ok && transientErr.Timeout() {
				isTransient = true
			}

			if isTransient {
				// For transient errors, set up retry sequence
				if tc.shouldSucceed {
					// If we expect success, set up the exact number of errors followed by success
					gomock.InOrder(
						// Return error tc.errorRetries times
						mockProvider.EXPECT().ResolveToolCall(
							gomock.Any(),
							gomock.Any(),
							gomock.Eq(tools),
						).Return(models.LanguageModelToolCallResolveOutput{}, tc.errorType).Times(tc.errorRetries),
						// Then return success
						mockProvider.EXPECT().ResolveToolCall(
							gomock.Any(),
							gomock.Any(),
							gomock.Eq(tools),
						).Return(models.LanguageModelToolCallResolveOutput{}, nil).Times(1),
					)

					mockProvider.EXPECT().StreamText(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
						emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, tc.expectedOutput))
					}).Return(models.LanguageModelOutput{Text: tc.expectedOutput, ModelName: modelName}, nil)
				} else {
					// If we expect failure (exceeds retry limit), allow multiple calls to fail
					mockProvider.EXPECT().ResolveToolCall(
						gomock.Any(),
						gomock.Any(),
						gomock.Eq(tools),
					).Return(models.LanguageModelToolCallResolveOutput{}, tc.errorType).MinTimes(tc.errorRetries)
				}
			} else {
				// For non-transient errors, just expect one call that fails
				mockProvider.EXPECT().ResolveToolCall(
					gomock.Any(),
					gomock.Any(),
					gomock.Eq(tools),
				).Return(models.LanguageModelToolCallResolveOutput{}, tc.errorType).Times(1)
			}

			output := StreamText(ctx, Params{
				Prompt:        tc.prompt,
				Provider:      mockProvider,
				EndConditions: []models.EndCondition{endconditions.NewMaxStepsEndCondition(10)},
				Tools:         tools,
			})

			var endPart models.ProcessEndPart
			var textParts []string
			for part := range output.Channel {
				switch p := part.(type) {
				case models.ProcessEndPart:
					endPart = p
				case models.TextPart:
					textParts = append(textParts, p.Text())
				}
			}

			if tc.shouldSucceed {
				assert.NotNil(t, endPart, "should have process end part for successful transient error recovery")
				assert.Contains(t, textParts, tc.expectedOutput, "should return expected output")
				assert.Equal(t, modelName, output.ModelName)
			} else {
				// For failures, we may or may not have an endPart depending on when error occurred
				assert.Equal(t, modelName, output.ModelName)
			}
		})
	}
}
