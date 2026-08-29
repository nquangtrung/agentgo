package fsm

import "context"

type EndState struct {
}

func (s *EndState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	// Implement the logic for end state here
	return nil, nil
}
