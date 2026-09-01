package fsm

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
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

type ExecutionContextAccumulator = func(acc *models.ToolExecutionsArchive, item *models.ToolExecuteOutput, messages *[]models.Message)

func (s *TextGenerationState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	messages := fsmCtx.ResolveCurrentStepMessages(ctx)

	output := resolveTextOutputAsToolExecuteOutput(
		provider.GenerateText(ctx, providers.AgentProviderPromptMessageParams{
			Messages: messages,
		}),
	)

	models.AccumulateToolCallResult(fsmCtx.ToolExecutionsArchive, &output, fsmCtx.Messages)
	return &AfterTextGenerationState{
		output: &output,
	}, nil
}
