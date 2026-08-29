package models

type EndCondition interface {
	Condition(context *ExecutionContext) bool
}

type MaxStepsEndCondition struct {
	MaxSteps int
}

func (m MaxStepsEndCondition) Condition(context *ExecutionContext) bool {
	return len(context.Steps()) >= m.MaxSteps
}

func NewMaxStepsEndCondition(maxSteps int) MaxStepsEndCondition {
	return MaxStepsEndCondition{
		MaxSteps: maxSteps,
	}
}
