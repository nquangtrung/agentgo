package fsm

import (
	"context"
	"log"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

type TextStreamState struct {
}

func (s *TextStreamState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)

	messages := fsmCtx.ResolveCurrentStepMessages(ctx)
	output := resolveTextOutputAsToolExecuteOutput(
		provider.StreamText(
			ctx,
			providers.AgentProviderPromptMessageParams{Messages: messages},
			*emitter,
		),
	)

	log.Printf("TextStreamState: Finished streaming text")
	models.AccumulateToolCallResult(fsmCtx.ToolExecutionsArchive, &output, fsmCtx.Messages)

	return &AfterTextGenerationState{
		output: &output,
	}, nil
}
