package agentgo

import (
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

type EndCondition interface {
	Condition(context *models.ExecutionContext) bool
}

type Params struct {
	Provider      providers.AgentProvider
	Prompt        string
	ModelName     string
	Messages      []models.Message
	Tools         []models.BaseTool
	EndConditions []EndCondition
}

type MaxStepsEndCondition struct {
	MaxSteps int
}

func (m MaxStepsEndCondition) Condition(context *models.ExecutionContext) bool {
	return len(context.Steps()) >= m.MaxSteps
}

func NewMaxStepsEndCondition(maxSteps int) MaxStepsEndCondition {
	return MaxStepsEndCondition{
		MaxSteps: maxSteps,
	}
}
