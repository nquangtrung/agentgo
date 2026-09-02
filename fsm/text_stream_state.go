package fsm

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

type TextStreamState struct {
}

func (s *TextStreamState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)

	messages := fsmCtx.ResolveCurrentStepMessages(ctx)
	output, err := provider.StreamText(
		ctx,
		providers.AgentProviderPromptMessageParams{Messages: messages},
		*emitter,
	)

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
