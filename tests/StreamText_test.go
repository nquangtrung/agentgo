package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"trontria.com/agentgo"
	"trontria.com/agentgo/mocks"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

func TestStreamText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Say this is a test"
	modelName := "mocked-llm-3.6-flash"
	// result := "This is a test"

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := agentgo.Params{
		Prompt:   prompt,
		Provider: mockProvider,
	}
	mockProvider.EXPECT().GetContext().Return(models.LanguageModelContext{
		ModelName: modelName,
	})

	mockProvider.EXPECT().StreamText(
		gomock.Cond(
			func(p providers.AgentProviderStreamTextParams) bool {
				if len(p.Messages) != 1 {
					return false
				}

				if p.Messages[0].GetContent().GetText() != params.Prompt {
					return false
				}

				return true
			}),
		gomock.Cond(
			func(channel chan models.Part) bool {
				return true
			},
		),
	).Do(func(params providers.AgentProviderStreamTextParams, channel chan models.Part) {
		context := models.LanguageModelContext{
			ModelName: modelName,
		}
		t.Log(channel)
		channel <- models.NewStepStartPart(context, "step start")
		channel <- models.NewStepEndPart(context, "step start", models.LanguageModelUsage{
			OutputTokens:    123,
			InputTokens:     245,
			CachedTokens:    21,
			ReasoningTokens: 12,
		})
	})
	output := agentgo.StreamText(params)
	if output == (models.LanguageModelStreamOutput{}) {
		panic("stream output is nil")
	}
	for part := range output.Channel {
		t.Logf("Part: %s", part.GetType())
		switch part.GetType() {
		case models.PartTypeStepStart:
			{
				p, ok := part.AsStepStartPart()
				assert.True(t, ok, "should be able to convert to step start")
				assert.NotNil(t, p, "should be able to access converted step")
			}
		case models.PartTypeStepEnd:
			{
				p, ok := part.AsStepEndPart()
				assert.True(t, ok, "should be able to convert to step end")
				assert.NotNil(t, p.GetUsage(), "should contain usage")
				// assert.Equal(t, p.GetUsage().InputTokens, 245, "input token should be correct")
			}
		}
	}
}
