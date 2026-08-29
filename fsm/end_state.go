package fsm

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"
)

type EndState struct {
}

func (s *EndState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	runner := ctx.Value(models.RunnerContextKey).(*utils.Runner[*models.ToolExecuteOutput, models.ExecutionContext])
	runner.Close()

	return nil, nil
}
