package fsm

import (
	"context"
)

type FinalizedState struct {
}

func (s *FinalizedState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	finalizeStep(ctx, fsmCtx)

	return &EndState{}, nil
}
