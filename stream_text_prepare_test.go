package agentgo

import (
	"context"
	"testing"

	"github.com/nquangtrung/agentgo/endconditions"
	"github.com/nquangtrung/agentgo/mocks"
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// ============================================================================
// PREPARE STEP OPTIONS TESTS
// ============================================================================

func TestStreamTextPrepareStepWithToolChoice(t *testing.T) {
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

	prepareStep := func(step step, ctx agentState) (PrepareStepResult, error) {
		// Only first step has tool choice override
		if step.index == 1 {
			return PrepareStepResult{
				ToolChoice: &ToolChoice{Name: "tool_a"},
			}, nil
		}
		return PrepareStepResult{}, nil
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	mockProvider.EXPECT().ResolveToolCall(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(models.LanguageModelToolCallResolveOutput{}, nil).MinTimes(1)
	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, "done"))
	}).Return(models.LanguageModelOutput{Text: "done", ModelName: modelName}, nil)

	output := StreamText(ctx, Params{
		Prompt:      prompt,
		Provider:    mockProvider,
		Tools:       tools,
		PrepareStep: prepareStep,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
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

func TestStreamTextPrepareStepWithMessages(t *testing.T) {
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

	prepareStep := func(step step, ctx agentState) (PrepareStepResult, error) {
		if step.index == 1 {
			// Override messages on first step
			customMessages := []models.Message{
				models.NewStringMessage("system", "You are a helpful assistant"),
				models.NewStringMessage("user", "Modified prompt in prepare step"),
			}
			return PrepareStepResult{
				Messages: &customMessages,
			}, nil
		}
		return PrepareStepResult{}, nil
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
		gomock.Any(),
	).Return(models.LanguageModelToolCallResolveOutput{}, nil).MinTimes(1)

	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, "done"))
	}).Return(models.LanguageModelOutput{Text: "done", ModelName: modelName}, nil)

	output := StreamText(ctx, Params{
		Prompt:      prompt,
		Provider:    mockProvider,
		Tools:       tools,
		PrepareStep: prepareStep,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	})

	var endPart models.ProcessEndPart
	for part := range output.Channel {
		if p, ok := part.(models.ProcessEndPart); ok {
			endPart = p
		}
	}

	assert.NotNil(t, endPart)
	assert.True(t, messageCheckPassed, "should have passed modified messages to ResolveToolCall")
}

func TestStreamTextPrepareStepWithActiveTools(t *testing.T) {
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

	prepareStep := func(step step, ctx agentState) (PrepareStepResult, error) {
		if step.index == 1 {
			// Restrict to only tool_1
			activeTools := []string{"tool_1"}
			return PrepareStepResult{
				ActiveTools: &activeTools,
			}, nil
		}
		return PrepareStepResult{}, nil
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

	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, "done"))
	}).Return(models.LanguageModelOutput{Text: "done", ModelName: modelName}, nil)

	output := StreamText(ctx, Params{
		Prompt:      prompt,
		Provider:    mockProvider,
		Tools:       tools,
		PrepareStep: prepareStep,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	})

	var endPart models.ProcessEndPart
	for part := range output.Channel {
		if p, ok := part.(models.ProcessEndPart); ok {
			endPart = p
		}
	}

	assert.NotNil(t, endPart)
}

func TestStreamTextPrepareStepMultipleOverrides(t *testing.T) {
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

	prepareStep := func(step step, ctx agentState) (PrepareStepResult, error) {
		if step.index == 1 {
			// Override all three options
			activeTools := []string{"search"}
			customMessages := []models.Message{
				models.NewStringMessage("user", "Search for information"),
			}
			return PrepareStepResult{
				ToolChoice:  &ToolChoice{Name: "search"},
				Messages:    &customMessages,
				ActiveTools: &activeTools,
			}, nil
		}
		return PrepareStepResult{}, nil
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
	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, "search completed"))
	}).Return(models.LanguageModelOutput{Text: "search completed", ModelName: modelName}, nil)

	output := StreamText(ctx, Params{
		Prompt:      prompt,
		Provider:    mockProvider,
		Tools:       tools,
		PrepareStep: prepareStep,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
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

func TestStreamTextPrepareStepPerStepOptions(t *testing.T) {
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

	prepareStep := func(step step, ctx agentState) (PrepareStepResult, error) {
		if step.index == 1 {
			// First step: use only step1_tool
			activeTools := []string{"step1_tool"}
			return PrepareStepResult{
				ActiveTools: &activeTools,
			}, nil
		} else if step.index == 2 {
			// Second step: use only step2_tool
			activeTools := []string{"step2_tool"}
			return PrepareStepResult{
				ActiveTools: &activeTools,
			}, nil
		}
		return PrepareStepResult{}, nil
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

	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, "multi-step complete"))
	}).Return(models.LanguageModelOutput{Text: "multi-step complete", ModelName: modelName}, nil)

	output := StreamText(ctx, Params{
		Prompt:      prompt,
		Provider:    mockProvider,
		Tools:       tools,
		PrepareStep: prepareStep,
		EndConditions: []models.EndCondition{
			endconditions.NewMaxStepsEndCondition(5),
		},
	})

	var endPart models.ProcessEndPart
	for part := range output.Channel {
		if p, ok := part.(models.ProcessEndPart); ok {
			endPart = p
		}
	}

	assert.NotNil(t, endPart)
}
