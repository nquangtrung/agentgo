package fsm

import "context"

type ErrorState struct {
	Err error
}

func (s *ErrorState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	return nil, s.Err
}
