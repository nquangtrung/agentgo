package models

type StepPart interface {
	StepName() string
}

type StepStartPart interface {
	Part
	StepPart
	StartPart
}

type StepPartImpl struct {
	BasePart
	stepName string
}

type BaseStepStartPart struct {
	StepPartImpl
	StartPartImpl
}

type StepEndPart interface {
	Part
	StepPart
	EndPart
}

type BaseStepEndPart struct {
	StepPartImpl
	EndPartImpl
}

func (s StepPartImpl) StepName() string {
	return s.stepName
}

func NewStepStartPart(context LanguageModelContext, stepName string) *BaseStepStartPart {
	return &BaseStepStartPart{
		StepPartImpl: StepPartImpl{
			BasePart: BasePart{
				partType: PartTypeStepStart,
				context:  context,
			},
			stepName: stepName,
		},
	}
}

func NewStepEndPart(context LanguageModelContext, stepName string, usage LanguageModelUsage) *BaseStepEndPart {
	return &BaseStepEndPart{
		StepPartImpl: StepPartImpl{
			BasePart: BasePart{
				partType: PartTypeStepEnd,
				context:  context,
			},
			stepName: stepName,
		},
		EndPartImpl: NewEndPart(usage, FinishReasonCompleted),
	}
}
