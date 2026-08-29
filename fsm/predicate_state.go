package fsm

import (
	"context"

	"trontria.com/agentgo/models"
)

type PredicateState struct {
}

func canProceedToNextStep(context *models.ExecutionContext, endConds []models.EndCondition) bool {
	conditions := endConds
	for _, condition := range conditions {
		if condition.Condition(context) {
			return false
		}
	}
	return true
}

func (s *PredicateState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	endConditions := ctx.Value(models.EndConditionsContextKey).([]models.EndCondition)
	tools := ctx.Value(models.ToolsContextKey).([]models.BaseTool)

	switch {
	case len(endConditions) == 0 || len(tools) == 0:
		return &TextGenerationState{}, nil
	case canProceedToNextStep(fsmCtx.ExecutionContext, endConditions):
		return &ToolResolveState{}, nil
	default:
		return &EndState{}, nil
	}
}
