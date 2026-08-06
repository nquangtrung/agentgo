package models

type StepPart interface {
	Part
	GetStepName() string
}

type StepStartPart interface {
	StepPart
}

type StepPartImpl struct {
	PartImpl
	StepName string `json:"step_name,omitempty"`
}

type StepStartPartImpl struct {
	StepPartImpl
}

type StepEndPart interface {
	StepPart
	EndPart
}

type StepEndPartImpl struct {
	StepPartImpl
	EndPartImpl
}

func (s StepPartImpl) GetStepName() string {
	return s.StepName
}

func NewStepStartPart(context LanguageModelContext, stepName string) StepStartPart {
	return StepStartPartImpl{
		StepPartImpl: StepPartImpl{
			PartImpl: NewPart(context, PartTypeStepStart),
			StepName: stepName,
		},
	}
}

func ToStepStartPart(part Part) (StepStartPart, bool) {
	if stepStartPart, ok := part.(StepStartPart); ok {
		return stepStartPart, true
	}
	return nil, false
}

func NewStepEndPart(context LanguageModelContext, stepName string, usage LanguageModelUsage) StepEndPart {
	return StepEndPartImpl{
		StepPartImpl: StepPartImpl{
			PartImpl: NewPart(context, PartTypeStepEnd),
			StepName: stepName,
		},
		EndPartImpl: NewEndPart(usage, FinishReasonCompleted),
	}
}

func ToStepEndPart(part Part) (StepEndPart, bool) {
	if stepEndPart, ok := part.(StepEndPart); ok {
		return stepEndPart, true
	}
	return nil, false
}
