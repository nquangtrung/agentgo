package fsm

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

type TextStreamState struct {
}

func (s *TextStreamState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	messages := fsmCtx.Messages
	partChannel := ctx.Value(models.StreamPartChannelContextKey).(chan models.Part)

	provider.StreamText(ctx, providers.AgentProviderPromptMessageParams{Messages: *messages}, partChannel)

	return &EndState{}, nil
}
