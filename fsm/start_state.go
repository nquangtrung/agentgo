package fsm

import "context"

type StartState struct {
}

func (s *StartState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	return &StepStartState{}, nil
}
