package fsm

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"
)

type StepStartState struct {
}

func (s *StepStartState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	runner := ctx.Value(models.RunnerContextKey).(*utils.Runner[*models.ToolExecuteOutput, models.ExecutionContext])
	runner.EmitStart(utils.RunnerEventStart, "step", fsmCtx.ExecutionContext)
	return &PredicateState{}, nil
}
