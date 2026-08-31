package fsm

import (
	"context"

	"trontria.com/agentgo/models"
)

type AfterTextGenerationState struct {
	output *models.ToolExecuteOutput
}

func (s *AfterTextGenerationState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	fsmCtx.TextGenerated = true

	if s.output == nil {
		// Should never happens
		panic("AfterTextGenerationState: output is nil")
	}

	fsmCtx.CurrentStep.Usage = models.AccumulateUsage(
		fsmCtx.CurrentStep.Usage,
		s.output.Usage,
	)

	return &FinalizedState{}, nil
}
