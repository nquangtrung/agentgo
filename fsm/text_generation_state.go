package fsm

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

type TextGenerationState struct {
}

func resolveTextOutputAsToolExecuteOutput(textOutput models.LanguageModelOutput, err error) models.ToolExecuteOutput {
	if err != nil {
		return models.ToolExecuteOutput{
			Output: nil,
			Error:  err,
			Usage:  textOutput.Usage,
		}
	}
	return models.ToolExecuteOutput{
		Output: map[string]any{
			"text": textOutput.Text,
		},
		Error: nil,
		Usage: textOutput.Usage,
	}
}
func (s *TextGenerationState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	runner := ctx.Value(models.RunnerContextKey).(*utils.Runner[*models.ToolExecuteOutput, models.ExecutionContext])
	messages := fsmCtx.Messages

	runner.Execute(
		"text_generation",
		func(acc *models.ExecutionContext) (*models.ToolExecuteOutput, error) {
			output := resolveTextOutputAsToolExecuteOutput(
				provider.GenerateText(ctx, providers.AgentProviderPromptMessageParams{
					Messages: *messages,
				}),
			)
			return &output, nil
		},
		fsmCtx.ExecutionContext,
	)

	return &EndState{}, nil
}
