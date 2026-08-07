package models

type StepPart interface {
	GetStepName() string
}

type StepStartPart interface {
	Part
	StepPart
}

type StepPartImpl struct {
	PartImpl
	stepName string
}

type StepStartPartImpl struct {
	StepPartImpl
}

type StepEndPart interface {
	Part
	StepPart
	EndPart
}

type StepEndPartImpl struct {
	StepPartImpl
	EndPartImpl
}

func (s StepPartImpl) GetStepName() string {
	return s.stepName
}

func NewStepStartPart(context LanguageModelContext, stepName string) StepStartPart {
	return StepStartPartImpl{
		StepPartImpl: StepPartImpl{
			PartImpl: NewPart(context, PartTypeStepStart),
			stepName: stepName,
		},
	}
}

func NewStepEndPart(context LanguageModelContext, stepName string, usage LanguageModelUsage) StepEndPart {
	return StepEndPartImpl{
		StepPartImpl: StepPartImpl{
			PartImpl: NewPart(context, PartTypeStepEnd),
			stepName: stepName,
		},
		EndPartImpl: NewEndPart(usage, FinishReasonCompleted),
	}
}
