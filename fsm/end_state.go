package fsm

import (
	"context"
)

type EndState struct {
}

func (s *EndState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	return nil, nil
}
