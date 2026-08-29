package fsm

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"
)

type StepEndState struct {
	toEnd bool
}

func (s *StepEndState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	runner := ctx.Value(models.RunnerContextKey).(*utils.Runner[*models.ToolExecuteOutput, models.ExecutionContext])
	runner.EmitSuccess(utils.RunnerEventSuccess, "step", &models.ToolExecuteOutput{}, fsmCtx.ExecutionContext)

	if !s.toEnd {
		return &StepStartState{}, nil
	} else {
		return &EndState{}, nil
	}
}
