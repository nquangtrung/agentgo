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
	"trontria.com/agentgo/mocks"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

func checkPart(channel <-chan models.Part, check func(part models.Part)) {
	part := <-channel
	log.Printf("Received part: %s", part.Type())
	check(part)
}

func TestStreamText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Say this is a test"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := Params{
		Prompt:   prompt,
		Provider: mockProvider,
	}
	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})

	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Cond(
			func(p providers.AgentProviderPromptMessageParams) bool {
				switch {
				case len(p.Messages) != 1:
					return false
				case p.Messages[0].Content().Text() != params.Prompt:
					return false
				default:
					return true
				}
			}),
		gomock.Any(),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		context := models.LanguageModelContext{
			ModelName: modelName,
		}
		for i := range 3 {
			emitter.Emit(models.NewTextPart(context, fmt.Sprintf("text %d", i)))
		}
	})
	output := StreamText(ctx, params)
	assert.NotNil(t, output.Channel, "Output channel should not be nil")
	assert.Equal(t, modelName, output.ModelName, "Output model name should be correct")

	var texts []string = []string{}

	checkPart(output.Channel, func(part models.Part) {
		_, ok := part.(models.ProcessStartPart)
		assert.True(t, ok, "part should be of type ProcessStartPart")
	})
	checkPart(output.Channel, func(part models.Part) {
		_, ok := part.(models.StepStartPart)
		assert.True(t, ok, "part should be of type StepStartPart")
	})
	for range 3 {
		checkPart(output.Channel, func(part models.Part) {
			textPart, ok := part.(models.TextPart)
			assert.True(t, ok, "part should be of type TextPart")
			texts = append(texts, textPart.Text())
		})
	}
	checkPart(output.Channel, func(part models.Part) {
		_, ok := part.(models.StepEndPart)
		assert.True(t, ok, "part should be of type StepEndPart")
	})
	checkPart(output.Channel, func(part models.Part) {
		_, ok := part.(models.ProcessEndPart)
		assert.True(t, ok, "part should be of type ProcessEndPart")
	})

	assert.ElementsMatch(
		t, []string{"text 0", "text 1", "text 2"}, texts, "should receive correct deltas",
	)
}

func TestStreamTextWithTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProvider := mocks.NewMockAgentProvider(ctrl)

	prompt := "Say this is a test"
	modelName := "mocked-llm-3.6-flash"
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
		returned []models.ToolCall,
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
			[]models.ToolCall{
				{
					ToolName: "mock_tool",
					Params:   toolParams,
				},
			}),
		mockResolveToolCallReturn(
			gomock.Cond(checkResolveToolCall),
			gomock.Eq(tools),
			nil,
		),
	)
	mockProvider.EXPECT().Context().MinTimes(1).Return(models.LanguageModelContext{
		ModelName: modelName,
	})

	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Cond(
			func(p providers.AgentProviderPromptMessageParams) bool {
				switch {
				case len(p.Messages) != 2:
					return false
				case p.Messages[0].Content().Text() != params.Prompt:
					return false
				case p.Messages[1].Content().Text() != fmt.Sprintf("Tool [%s] execution result: %s", tools[0].Name(), jsonedToolResult):
					log.Printf("Expected second message to be tool result, got '%s'", p.Messages[1].Content().Text())
					return false
				default:
					return true
				}
			}),
		gomock.Any(),
	).DoAndReturn(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) (models.LanguageModelOutput, error) {
		log.Printf("StreamText called with params: %v", params)
		context := models.LanguageModelContext{
			ModelName: modelName,
		}
		for i := range 3 {
			emitter.Emit(models.NewTextPart(context, fmt.Sprintf("text %d", i)))
		}
		return models.LanguageModelOutput{
			Text: "text 0text 1text 2",
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
		}, nil
	})

	output := StreamText(ctx, params)
	assert.NotNil(t, output.Channel, "Output channel should not be nil")
	assert.Equal(t, modelName, output.ModelName, "Output model name should be correct")

	var endPart models.ProcessEndPart
	var texts []string
	for part := range output.Channel {
		log.Printf("Received part: %s", part.Type())
		switch p := part.(type) {
		case models.ProcessEndPart:
			endPart = p
		case models.TextPart:
			texts = append(texts, p.Text())
		}
	}

	assert.NotNil(t, endPart, "process end part should be emitted")
	assert.Equal(t, modelName, output.ModelName, "should have correct model name")
	assert.ElementsMatch(t, []string{"text 0", "text 1", "text 2"}, texts, "should receive correct deltas")
	assert.Equal(t, int64(245+200), endPart.Usage().InputTokens, "should have correct input tokens")
	assert.Equal(t, int64(123+100), endPart.Usage().OutputTokens, "should have correct output tokens")
	assert.Equal(t, int64(21+20), endPart.Usage().InputTokensDetails.CachedTokens, "should have correct cached token")
	assert.Equal(t, int64(12+0), endPart.Usage().OutputTokensDetails.ReasoningTokens, "should have correct reasoning token")
}

func TestStreamTextStopsAtMaxSteps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	prompt := "Stop after one tool call"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()
	tools := []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "mock_tool",
			Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
				assert.Equal(t, map[string]any{"param": "value"}, params.Input)
				return models.ToolExecuteOutput{
					Output: map[string]any{"status": "ok"},
					Usage: models.LanguageModelUsage{
						OutputTokens: 40,
						InputTokens:  50,
						InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
							CachedTokens: 5,
						},
						OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
							ReasoningTokens: 4,
						},
						TotalTokens: 90,
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
	).Return([]models.ToolCall{{ToolName: "mock_tool", Params: map[string]any{"param": "value"}}}, nil)

	output := StreamText(ctx, Params{
		Prompt:        prompt,
		Provider:      mockProvider,
		EndConditions: []models.EndCondition{endconditions.NewMaxStepsEndCondition(1)},
		Tools:         tools,
	})

	var endPart models.ProcessEndPart
	for part := range output.Channel {
		if p, ok := part.(models.ProcessEndPart); ok {
			endPart = p
		}
	}

	assert.Equal(t, modelName, output.ModelName)
	assert.NotNil(t, endPart)
	assert.Equal(t, int64(50), endPart.Usage().InputTokens)
	assert.Equal(t, int64(40), endPart.Usage().OutputTokens)
	assert.Equal(t, int64(5), endPart.Usage().InputTokensDetails.CachedTokens)
	assert.Equal(t, int64(4), endPart.Usage().OutputTokensDetails.ReasoningTokens)
}

func TestStreamTextMultipleToolCalls(t *testing.T) {
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
	finalUsage := models.LanguageModelUsage{OutputTokens: 35, InputTokens: 40, InputTokensDetails: models.LanguageModelUsageInputTokensDetails{CachedTokens: 4}, OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{ReasoningTokens: 3}, TotalTokens: 75}
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
		).Return([]models.ToolCall{{ToolName: "first_tool", Params: tool1Params}, {ToolName: "second_tool", Params: tool2Params}}, nil),
		mockProvider.EXPECT().ResolveToolCall(
			gomock.Any(),
			gomock.Any(),
			gomock.Eq(tools),
		).Return(nil, nil),
	)
	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, "done"))
	}).Return(models.LanguageModelOutput{Text: "done", Usage: finalUsage, ModelName: modelName}, nil)

	output := StreamText(ctx, Params{
		Prompt:        prompt,
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

	assert.Equal(t, modelName, output.ModelName)
	assert.NotNil(t, endPart)
	assert.Equal(t, []string{"done"}, textParts)
	assert.Equal(t, int64(20+30+40), endPart.Usage().InputTokens)
	assert.Equal(t, int64(10+25+35), endPart.Usage().OutputTokens)
	assert.Equal(t, int64(2+3+4), endPart.Usage().InputTokensDetails.CachedTokens)
	assert.Equal(t, int64(1+2+3), endPart.Usage().OutputTokensDetails.ReasoningTokens)
}
