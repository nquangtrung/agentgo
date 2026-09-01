package fsm

import (
	"context"

	"trontria.com/agentgo/models"
)

type PredicateState struct {
}

func canProceedToNextStep(context *models.ToolExecutionsArchive, endConds []models.EndCondition) bool {
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

	step := fsmCtx.CurrentStep
	if step.PrepareStepResult.ToolChoice != nil {
		toolChoice := step.PrepareStepResult.ToolChoice
		if toolChoice.Name != "" {
			return &ToolResolveState{}, nil
		}
	}

	switch {
	case fsmCtx.TextGenerated:
		return &FinalizedState{}, nil
	case len(endConditions) == 0 || len(tools) == 0:
		return &PrepareTextGenerationState{}, nil
	case canProceedToNextStep(fsmCtx.ToolExecutionsArchive, endConditions):
		return &ToolResolveState{}, nil
	default:
		return &FinalizedState{}, nil
	}
}
