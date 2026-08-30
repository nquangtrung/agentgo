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
	messages := fsmCtx.Messages
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)

	output := resolveTextOutputAsToolExecuteOutput(
		provider.StreamText(
			ctx,
			providers.AgentProviderPromptMessageParams{Messages: *messages},
			*emitter,
		),
	)

	log.Printf("TextStreamState: Finished streaming text")
	models.AccumulateToolCallResult(fsmCtx.ToolExecutionsArchive, &output, messages)

	return &AfterTextGenerationState{
		output: &output,
	}, nil
}
