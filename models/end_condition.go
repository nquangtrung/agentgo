package models

type EndCondition interface {
	Condition(context *ToolExecutionsArchive) bool
}

type MaxStepsEndCondition struct {
	MaxSteps int
}

func (m MaxStepsEndCondition) Condition(context *ToolExecutionsArchive) bool {
	return len(context.Records()) >= m.MaxSteps
}

func NewMaxStepsEndCondition(maxSteps int) MaxStepsEndCondition {
	return MaxStepsEndCondition{
		MaxSteps: maxSteps,
	}
}
