package agentgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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
						OutputTokens:    100,
						InputTokens:     200,
						CachedTokens:    20,
						ReasoningTokens: 0,
					},
				}
			},
		}),
	}
	params := Params{
		Prompt:   prompt,
		Provider: mockProvider,
		EndConditions: []models.EndCondition{
			models.NewMaxStepsEndCondition(5),
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
		gomock.Cond(
			func(channel chan models.Part) bool {
				return true
			},
		),
	).Do(func(ctx context.Context, params providers.AgentProviderPromptMessageParams, channel chan models.Part) {
		log.Printf("StreamText called with params: %v", params)
		context := models.LanguageModelContext{
			ModelName: modelName,
		}
		for i := range 3 {
			channel <- models.NewTextPart(context, fmt.Sprintf("text %d", i))
		}
	})

	output := StreamText(ctx, params)
	assert.NotNil(t, output.Channel, "Output channel should not be nil")
	assert.Equal(t, modelName, output.ModelName, "Output model name should be correct")

	for part := range output.Channel {
		log.Printf("Received part: %s", part.Type())
	}

	assert.Equal(t, modelName, output.ModelName, "should have correct model name")
	// assert.Equal(t, result, output.Text, "should have correct output")
	// assert.Equal(t, int64(245+200), output.Usage.InputTokens, "should have correct input tokens")
	// assert.Equal(t, int64(123+100), output.Usage.OutputTokens, "should have correct output tokens")
	// assert.Equal(t, int64(21+20), output.Usage.CachedTokens, "should have correct cached token")
	// assert.Equal(t, int64(12+0), output.Usage.ReasoningTokens, "should have correct reasoning token")
	// assert.Equal(t, 2, len(output.Context.Steps()), "should have correct number of steps")
	// assert.Equal(t, "mock_tool", utils.Must(output.Context.Step(0)).Name, "should have correct first step name")
	// assert.Equal(t, toolResult, utils.Must(output.Context.Step(0)).ToolResult.Output, "should have correct result in first step")
}
