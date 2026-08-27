package agentgo

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"trontria.com/agentgo/mocks"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

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
	mockProvider.EXPECT().Context().Return(models.LanguageModelContext{
		ModelName: modelName,
	})

	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Cond(
			func(p providers.AgentProviderPromptMessageParams) bool {
				if len(p.Messages) != 1 {
					return false
				}

				if p.Messages[0].Content().Text() != params.Prompt {
					return false
				}

				return true
			}),
		gomock.Cond(
			func(channel chan models.Part) bool {
				return true
			},
		),
	).Do(func(params providers.AgentProviderPromptMessageParams, channel chan models.Part) {
		context := models.LanguageModelContext{
			ModelName: modelName,
		}
		t.Log(channel)
		channel <- models.NewStepStartPart(context, "step start")
		for i := range 3 {
			channel <- models.NewTextPart(context, fmt.Sprintf("text %d", i))
		}
		channel <- models.NewStepEndPart(context, "step start", models.LanguageModelUsage{
			OutputTokens:    123,
			InputTokens:     245,
			CachedTokens:    21,
			ReasoningTokens: 12,
		})
	})
	output := StreamText(ctx, params)
	if output == (models.LanguageModelStreamOutput{}) {
		panic("stream output is nil")
	}

	var texts []string = []string{}
	for part := range output.Channel {
		switch p := part.(type) {
		case models.StepStartPart:
			{
				assert.NotNil(t, p, "should be able to access converted step")
			}
		case models.TextPart:
			{
				texts = append(texts, p.Text())
			}
		case models.StepEndPart:
			{
				assert.NotNil(t, p.Usage(), "should contain usage")
				assert.Equal(t, int64(245), p.Usage().InputTokens, "input token should be correct")
			}
		}
	}
	assert.ElementsMatch(
		t, texts, []string{"text 0", "text 1", "text 2"}, "should receive correct deltas",
	)
}
