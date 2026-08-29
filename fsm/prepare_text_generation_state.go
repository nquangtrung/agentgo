package fsm

import (
	"context"

	"trontria.com/agentgo/models"
)

type PrepareTextGenerationState struct {
}

func (s *PrepareTextGenerationState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	stream := ctx.Value(models.StreamContextKey).(bool)

	if stream {
		return &TextStreamState{}, nil
	} else {
		return &TextGenerationState{}, nil
	}
}
