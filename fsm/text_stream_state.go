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
	partChannel := ctx.Value(models.StreamPartChannelContextKey).(chan models.Part)

	log.Printf("TextStreamState: Executing with %d messages", len(*messages))
	log.Printf("TextStreamState: Provider %+v", provider)
	provider.StreamText(
		ctx,
		providers.AgentProviderPromptMessageParams{Messages: *messages},
		partChannel,
	)
	log.Printf("TextStreamState: Finished streaming text")

	return &AfterTextGenerationState{}, nil
}
