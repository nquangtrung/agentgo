package fsm

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

type TextGenerationState struct {
}

func resolveTextOutputAsToolExecuteOutput(textOutput models.LanguageModelOutput) models.ToolExecuteOutput {
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

	output, err := provider.GenerateText(ctx, providers.AgentProviderPromptMessageParams{
		Messages: messages,
	})

	switch {
	case err == nil:
		outputAsToolExecuteOutput := resolveTextOutputAsToolExecuteOutput(output)
		models.AccumulateToolCallResult(fsmCtx.ToolExecutionsArchive, &outputAsToolExecuteOutput, fsmCtx.Messages)
		return &AfterTextGenerationState{
			output: &outputAsToolExecuteOutput,
		}, nil
	case utils.IsTransientError(err):
		return &RetryState{
			OriginalState: s,
		}, nil
	default:
		// go to error state, and end the loop
		return &ErrorState{Err: err}, nil
	}
}
