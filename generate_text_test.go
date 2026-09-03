package agentgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"trontria.com/agentgo/endconditions"
	"trontria.com/agentgo/fsm"
	"trontria.com/agentgo/mocks"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

func TestGenerateText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Say this is a test"
	modelName := "mocked-llm-3.6-flash"
	result := "This is a test"
	ctx := context.Background()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := Params{
		Prompt:   prompt,
		Provider: mockProvider,
	}
	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Cond(
			func(p providers.AgentProviderPromptMessageParams) bool {
				switch {
				case len(p.Messages) != 1:
					log.Printf("Expected 1 message, got %d", len(p.Messages))
					return false
				case p.Messages[0].Content().Text() != params.Prompt:
					log.Printf("Expected prompt '%s', got '%s'", params.Prompt, p.Messages[0].Content().Text())
					return false
				default:
					return true
				}
			}),
	).Return(
		models.LanguageModelOutput{
			Text: result,
			Usage: models.LanguageModelUsage{
				OutputTokens: 123,
				InputTokens:  245,
				InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
					CachedTokens: 21,
				},
				OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
					ReasoningTokens: 12,
				},
				TotalTokens: 368,
			},
			ModelName: modelName,
		},
		nil,
	)

	output, err := GenerateText(ctx, params)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, output.ModelName, modelName, "should have correct model name")
	assert.Equal(t, output.Text, result, "should have correct output")
	assert.Equal(t, int64(245), output.Usage.InputTokens, "should have correct input tokens")
	assert.Equal(t, int64(123), output.Usage.OutputTokens, "should have correct output tokens")
	assert.Equal(t, int64(21), output.Usage.InputTokensDetails.CachedTokens, "should have correct cached token")
	assert.Equal(t, int64(12), output.Usage.OutputTokensDetails.ReasoningTokens, "should have correct reasoning token")
}

func TestGenerateTextWithTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProvider := mocks.NewMockAgentProvider(ctrl)

	prompt := "Say this is a test"
	modelName := "mocked-llm-3.6-flash"
	result := "This is a test"
	ctx := context.Background()

	toolParams := map[string]any{
		"param1": "value1",
		"param2": "value2",
	}
	toolResult := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
	jsonedToolResult := utils.Must(json.Marshal(toolResult))
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "mock_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				log.Printf("Tool called with params: %v", params)
				assert.Equal(t, toolParams, params.Input, "should be called with correct tool params")
				return models.ToolExecuteOutput{
					Output: toolResult,
					Usage: models.LanguageModelUsage{
						OutputTokens: 100,
						InputTokens:  200,
						InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
							CachedTokens: 20,
						},
						TotalTokens: 300,
					},
				}
			},
		}),
	}
	params := Params{
		Prompt:   prompt,
		Provider: mockProvider,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
		Tools: tools,
	}

	checkResolveToolCall := func(p providers.AgentProviderPromptMessageParams) bool {
		switch {
		case len(p.Messages) != 2:
			log.Printf("Expected at least 2 message, got %d", len(p.Messages))
			return false
		case p.Messages[0].Content().Text() != prompt:
			log.Printf("Expected first message to be prompt '%s', got '%s'", prompt, p.Messages[0].Content().Text())
			return false
		case p.Messages[1].Content().Text() != fmt.Sprintf("Tool [%s] execution result: %s", tools[0].Name(), jsonedToolResult):
			log.Printf("Expected second message to be tool result, got '%s'", p.Messages[1].Content().Text())
			return false
		default:
			return true
		}
	}

	mockResolveToolCallReturn := func(
		paramsMatcher gomock.Matcher,
		toolParamsMatcher gomock.Matcher,
		returned models.LanguageModelToolCallResolveOutput,
	) *gomock.Call {
		return mockProvider.EXPECT().ResolveToolCall(gomock.Any(), paramsMatcher, toolParamsMatcher).Return(returned, nil)
	}
	gomock.InOrder(
		mockResolveToolCallReturn(
			gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
				switch {
				case len(p.Messages) != 1:
					log.Printf("Expected 1 message, got %d", len(p.Messages))
					return false
				case p.Messages[0].Content().Text() != params.Prompt:
					log.Printf("Expected prompt '%s', got '%s'", params.Prompt, p.Messages[0].Content().Text())
					return false
				default:
					return true
				}
			}),
			gomock.Eq(tools),
			models.LanguageModelToolCallResolveOutput{
				ToolCalls: []models.ToolCall{
					{
						ToolName: "mock_tool",
						Params:   toolParams,
					},
				},
			}),
		mockResolveToolCallReturn(
			gomock.Cond(checkResolveToolCall),
			gomock.Eq(tools),
			models.LanguageModelToolCallResolveOutput{},
		),
	)
	mockProvider.EXPECT().Context().MinTimes(1).Return(models.LanguageModelContext{
		ModelName: modelName,
	})
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Cond(checkResolveToolCall),
	).Return(
		models.LanguageModelOutput{
			Text: result,
			Usage: models.LanguageModelUsage{
				OutputTokens: 123,
				InputTokens:  245,
				InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
					CachedTokens: 21,
				},
				OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
					ReasoningTokens: 12,
				},
				TotalTokens: 368,
			},
			ModelName: modelName,
		},
		nil,
	)
	output, err := GenerateText(ctx, params)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, modelName, output.ModelName, "should have correct model name")
	assert.Equal(t, result, output.Text, "should have correct output")
	assert.Equal(t, int64(245+200), output.Usage.InputTokens, "should have correct input tokens")
	assert.Equal(t, int64(123+100), output.Usage.OutputTokens, "should have correct output tokens")
	assert.Equal(t, int64(21+20), output.Usage.InputTokensDetails.CachedTokens, "should have correct cached token")
	assert.Equal(t, int64(12+0), output.Usage.OutputTokensDetails.ReasoningTokens, "should have correct reasoning token")
	assert.Equal(t, 2, len(output.Context.Records()), "should have correct number of steps")
	assert.Equal(t, "mock_tool", utils.Must(output.Context.Record(0)).Name, "should have correct first step name")
	assert.Equal(t, toolResult, utils.Must(output.Context.Record(0)).ToolResult.Output, "should have correct result in first step")
}

func TestGenerateTextStopsAtMaxSteps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Stop after one tool call"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()
	toolParams := map[string]any{"param": "value"}
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "mock_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				assert.Equal(t, toolParams, params.Input, "should be called with correct tool params")
				return models.ToolExecuteOutput{
					Output: map[string]any{"status": "ok"},
					Usage: models.LanguageModelUsage{
						OutputTokens: 50,
						InputTokens:  80,
						InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
							CachedTokens: 10,
						},
						OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
							ReasoningTokens: 5,
						},
						TotalTokens: 130,
					},
				}
			},
		}),
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Eq(tools),
	).Return(models.LanguageModelToolCallResolveOutput{
		ToolCalls: []models.ToolCall{{ToolName: "mock_tool", Params: toolParams}},
	}, nil)

	output, err := GenerateText(ctx, Params{
		Prompt:   prompt,
		Provider: mockProvider,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(1),
		},
		Tools: tools,
	})
	assert.NoError(t, err)
	assert.Equal(t, modelName, output.ModelName, "should have correct model name")
	assert.Empty(t, output.Text, "should stop before emitting final text")
	assert.Equal(t, int64(80), output.Usage.InputTokens, "should only include tool usage before max step")
	assert.Equal(t, int64(50), output.Usage.OutputTokens, "should only include tool usage before max step")
	assert.Equal(t, int64(10), output.Usage.InputTokensDetails.CachedTokens, "should only include tool usage before max step")
	assert.Equal(t, int64(5), output.Usage.OutputTokensDetails.ReasoningTokens, "should only include tool usage before max step")
	assert.Len(t, output.Context.Records(), 1, "should stop after a single tool call")
}

func TestGenerateTextMultipleToolCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Run both tools"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()
	tool1Params := map[string]any{"first": "value1"}
	tool2Params := map[string]any{"second": "value2"}
	tool1Result := map[string]any{"result": "first"}
	tool2Result := map[string]any{"result": "second"}
	tool1Usage := models.LanguageModelUsage{OutputTokens: 10, InputTokens: 20, InputTokensDetails: models.LanguageModelUsageInputTokensDetails{CachedTokens: 2}, OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{ReasoningTokens: 1}, TotalTokens: 30}
	tool2Usage := models.LanguageModelUsage{OutputTokens: 25, InputTokens: 30, InputTokensDetails: models.LanguageModelUsageInputTokensDetails{CachedTokens: 3}, OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{ReasoningTokens: 2}, TotalTokens: 55}
	finalUsage := models.LanguageModelUsage{OutputTokens: 40, InputTokens: 60, InputTokensDetails: models.LanguageModelUsageInputTokensDetails{CachedTokens: 4}, OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{ReasoningTokens: 3}, TotalTokens: 100}
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{Name: "first_tool", Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
			assert.Equal(t, tool1Params, params.Input)
			return models.ToolExecuteOutput{Output: tool1Result, Usage: tool1Usage}
		}}),
		models.NewTool(models.NewToolParams{Name: "second_tool", Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
			assert.Equal(t, tool2Params, params.Input)
			return models.ToolExecuteOutput{Output: tool2Result, Usage: tool2Usage}
		}}),
	}

	gomock.InOrder(
		mockProvider.EXPECT().ResolveToolCall(
			gomock.Any(),
			gomock.Any(),
			gomock.Eq(tools),
		).Return(models.LanguageModelToolCallResolveOutput{
			ToolCalls: []models.ToolCall{{ToolName: "first_tool", Params: tool1Params}, {ToolName: "second_tool", Params: tool2Params}},
		}, nil),
		mockProvider.EXPECT().ResolveToolCall(
			gomock.Any(),
			gomock.Any(),
			gomock.Eq(tools),
		).Return(models.LanguageModelToolCallResolveOutput{}, nil),
	)
	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelOutput{Text: "done", Usage: finalUsage, ModelName: modelName}, nil)

	output, err := GenerateText(ctx, Params{
		Prompt:        prompt,
		Provider:      mockProvider,
		EndConditions: []models.EndCondition{endconditions.NewMaxStepsEndCondition(10)},
		Tools:         tools,
	})
	assert.NoError(t, err)
	assert.Equal(t, "done", output.Text)
	assert.Equal(t, int64(20+30+60), output.Usage.InputTokens)
	assert.Equal(t, int64(10+25+40), output.Usage.OutputTokens)
	assert.Equal(t, int64(2+3+4), output.Usage.InputTokensDetails.CachedTokens)
	assert.Equal(t, int64(1+2+3), output.Usage.OutputTokensDetails.ReasoningTokens)
	assert.Len(t, output.Context.Records(), 3)
	// Tools are executed concurrently, so order is non-deterministic
	toolNames := []string{
		utils.Must(output.Context.Record(0)).ToolCalled.ToolName,
		utils.Must(output.Context.Record(1)).ToolCalled.ToolName,
	}
	assert.ElementsMatch(t, []string{"first_tool", "second_tool"}, toolNames)
	assert.Equal(t, "text", utils.Must(output.Context.Record(2)).Name)
}

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

// ============================================================================
// PREPARE STEP OPTIONS TESTS
// ============================================================================

func TestGenerateTextPrepareStepWithToolChoice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Test prepare step with tool choice"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "tool_a",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"called": "tool_a"},
					Usage:  models.LanguageModelUsage{OutputTokens: 10, InputTokens: 20, TotalTokens: 30},
				}
			},
		}),
		models.NewTool(models.NewToolParams{
			Name: "tool_b",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"called": "tool_b"},
					Usage:  models.LanguageModelUsage{OutputTokens: 10, InputTokens: 20, TotalTokens: 30},
				}
			},
		}),
	}

	stepCount := 0
	prepareStep := func(step fsm.Step, ctx fsm.AgentContext) (fsm.PrepareStepResult, error) {
		// Only first step has tool choice override
		if step.StepIndex == 1 {
			return fsm.PrepareStepResult{
				ToolChoice: &fsm.ToolChoice{Name: "tool_a"},
			}, nil
		}
		return fsm.PrepareStepResult{}, nil
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, tools []models.BaseTool) {
		stepCount++
	}).Return(models.LanguageModelToolCallResolveOutput{}, nil).MinTimes(1)
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelOutput{Text: "done", ModelName: modelName}, nil)

	output, err := GenerateText(ctx, Params{
		Prompt:      prompt,
		Provider:    mockProvider,
		Tools:       tools,
		PrepareStep: prepareStep,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, modelName, output.ModelName)
	assert.Equal(t, "done", output.Text)
}

func TestGenerateTextPrepareStepWithMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Original prompt"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "dummy_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"ok": true},
					Usage:  models.LanguageModelUsage{OutputTokens: 10, InputTokens: 20, TotalTokens: 30},
				}
			},
		}),
	}

	prepareStep := func(step fsm.Step, ctx fsm.AgentContext) (fsm.PrepareStepResult, error) {
		if step.StepIndex == 1 {
			// Override messages on first step
			customMessages := []models.Message{
				models.NewStringMessage("system", "You are a helpful assistant"),
				models.NewStringMessage("user", "Modified prompt in prepare step"),
			}
			return fsm.PrepareStepResult{
				Messages: &customMessages,
			}, nil
		}
		return fsm.PrepareStepResult{}, nil
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	
	// Track which messages are passed to ResolveToolCall
	messageCheckPassed := false
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
			// First call should have modified messages
			if len(p.Messages) == 2 && p.Messages[1].Content().Text() == "Modified prompt in prepare step" {
				messageCheckPassed = true
			}
			return true
		}),
		gomock.Eq(tools),
	).Return(models.LanguageModelToolCallResolveOutput{}, nil).MinTimes(1)

	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelOutput{Text: "done", ModelName: modelName}, nil)

	output, err := GenerateText(ctx, Params{
		Prompt:      prompt,
		Provider:    mockProvider,
		Tools:       tools,
		PrepareStep: prepareStep,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "done", output.Text)
	assert.True(t, messageCheckPassed, "should have passed modified messages to ResolveToolCall")
}

func TestGenerateTextPrepareStepWithActiveTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Test with active tools override"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "tool_1",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"ok": true},
					Usage:  models.LanguageModelUsage{OutputTokens: 10, InputTokens: 20, TotalTokens: 30},
				}
			},
		}),
		models.NewTool(models.NewToolParams{
			Name: "tool_2",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"ok": true},
					Usage:  models.LanguageModelUsage{OutputTokens: 10, InputTokens: 20, TotalTokens: 30},
				}
			},
		}),
	}

	prepareStep := func(step fsm.Step, ctx fsm.AgentContext) (fsm.PrepareStepResult, error) {
		if step.StepIndex == 1 {
			// Restrict to only tool_1
			activeTools := []string{"tool_1"}
			return fsm.PrepareStepResult{
				ActiveTools: &activeTools,
			}, nil
		}
		return fsm.PrepareStepResult{}, nil
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	
	// Verify that only tool_1 is passed
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Cond(func(tools []models.BaseTool) bool {
			// First call should have only tool_1
			return len(tools) == 1 && tools[0].Name() == "tool_1"
		}),
	).Return(models.LanguageModelToolCallResolveOutput{}, nil).Times(1)
	
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelToolCallResolveOutput{}, nil).MinTimes(0)

	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelOutput{Text: "done", ModelName: modelName}, nil)

	output, err := GenerateText(ctx, Params{
		Prompt:      prompt,
		Provider:    mockProvider,
		Tools:       tools,
		PrepareStep: prepareStep,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "done", output.Text)
}

func TestGenerateTextPrepareStepMultipleOverrides(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Test multiple prepare step overrides"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "search",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"results": []string{"result1", "result2"}},
					Usage:  models.LanguageModelUsage{OutputTokens: 20, InputTokens: 40, TotalTokens: 60},
				}
			},
		}),
		models.NewTool(models.NewToolParams{
			Name: "calculate",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"result": 42},
					Usage:  models.LanguageModelUsage{OutputTokens: 15, InputTokens: 30, TotalTokens: 45},
				}
			},
		}),
	}

	prepareStep := func(step fsm.Step, ctx fsm.AgentContext) (fsm.PrepareStepResult, error) {
		if step.StepIndex == 1 {
			// Override all three options
			activeTools := []string{"search"}
			customMessages := []models.Message{
				models.NewStringMessage("user", "Search for information"),
			}
			return fsm.PrepareStepResult{
				ToolChoice:  &fsm.ToolChoice{Name: "search"},
				Messages:    &customMessages,
				ActiveTools: &activeTools,
			}, nil
		}
		return fsm.PrepareStepResult{}, nil
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Cond(func(tools []models.BaseTool) bool {
			// Should only have search tool
			return len(tools) == 1 && tools[0].Name() == "search"
		}),
	).Return(models.LanguageModelToolCallResolveOutput{
		ToolCalls: []models.ToolCall{{ToolName: "search", Params: map[string]any{}}},
	}, nil)
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelToolCallResolveOutput{}, nil).MinTimes(0)
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelOutput{Text: "search completed", ModelName: modelName}, nil)

	output, err := GenerateText(ctx, Params{
		Prompt:      prompt,
		Provider:    mockProvider,
		Tools:       tools,
		PrepareStep: prepareStep,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "search completed", output.Text)
	assert.Equal(t, modelName, output.ModelName)
}

func TestGenerateTextPrepareStepPerStepOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Multi-step with different options"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "step1_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"step": "1"},
					Usage:  models.LanguageModelUsage{OutputTokens: 10, InputTokens: 20, TotalTokens: 30},
				}
			},
		}),
		models.NewTool(models.NewToolParams{
			Name: "step2_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				return models.ToolExecuteOutput{
					Output: map[string]any{"step": "2"},
					Usage:  models.LanguageModelUsage{OutputTokens: 10, InputTokens: 20, TotalTokens: 30},
				}
			},
		}),
	}

	callOrder := []string{}
	prepareStep := func(step fsm.Step, ctx fsm.AgentContext) (fsm.PrepareStepResult, error) {
		if step.StepIndex == 1 {
			// First step: use only step1_tool
			activeTools := []string{"step1_tool"}
			callOrder = append(callOrder, "step1")
			return fsm.PrepareStepResult{
				ActiveTools: &activeTools,
			}, nil
		} else if step.StepIndex == 2 {
			// Second step: use only step2_tool
			activeTools := []string{"step2_tool"}
			callOrder = append(callOrder, "step2")
			return fsm.PrepareStepResult{
				ActiveTools: &activeTools,
			}, nil
		}
		return fsm.PrepareStepResult{}, nil
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	
	// First ResolveToolCall should have step1_tool
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Cond(func(tools []models.BaseTool) bool {
			return len(tools) == 1 && tools[0].Name() == "step1_tool"
		}),
	).Return(models.LanguageModelToolCallResolveOutput{
		ToolCalls: []models.ToolCall{{ToolName: "step1_tool", Params: map[string]any{}}},
	}, nil).Times(1)
	
	// Second ResolveToolCall should have step2_tool
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Cond(func(tools []models.BaseTool) bool {
			return len(tools) == 1 && tools[0].Name() == "step2_tool"
		}),
	).Return(models.LanguageModelToolCallResolveOutput{
		ToolCalls: []models.ToolCall{{ToolName: "step2_tool", Params: map[string]any{}}},
	}, nil).Times(1)
	
	// Final ResolveToolCall
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelToolCallResolveOutput{}, nil).MinTimes(0)

	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelOutput{Text: "multi-step complete", ModelName: modelName}, nil)

	output, err := GenerateText(ctx, Params{
		Prompt:      prompt,
		Provider:    mockProvider,
		Tools:       tools,
		PrepareStep: prepareStep,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "multi-step complete", output.Text)
	assert.GreaterOrEqual(t, len(callOrder), 2, "should have called prepare step at least twice")
}
