package fsm

import (
	"context"
)

type AfterTextGenerationState struct {
}

func (s *AfterTextGenerationState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	fsmCtx.TextGenerated = true
	return &StepEndState{toEnd: true}, nil
}
