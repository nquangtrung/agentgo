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
	BaseStartPart
}

type StepEndPart interface {
	Part
	StepPart
	EndPart
}

type BaseStepEndPart struct {
	StepPartImpl
	BaseEndPart
}

func (s StepPartImpl) StepName() string {
	return s.stepName
}

func NewStepStartPart(context LanguageModelContext, stepName string) *BaseStepStartPart {
	return &BaseStepStartPart{
		StepPartImpl: StepPartImpl{
			BasePart: BasePart{partType: PartTypeStepStart, context: context},
			stepName: stepName,
		},
	}
}

func NewStepEndPart(context LanguageModelContext, stepName string, usage LanguageModelUsage) *BaseStepEndPart {
	return &BaseStepEndPart{
		StepPartImpl: StepPartImpl{
			BasePart: BasePart{partType: PartTypeStepEnd, context: context},
			stepName: stepName,
		},
		BaseEndPart: NewEndPart(usage, FinishReasonCompleted),
	}
}

type StepErrorPart interface {
	Part
	StepPart
	EndPart
	Error() error
}

type BaseStepErrorPart struct {
	StepPartImpl
	BaseEndPart
	error error
}

func (s BaseStepErrorPart) Error() error {
	return s.error
}

func NewStepErrorPart(context LanguageModelContext, stepName string, usage LanguageModelUsage, err error) *BaseStepErrorPart {
	return &BaseStepErrorPart{
		StepPartImpl: StepPartImpl{
			BasePart: BasePart{partType: PartTypeStepError, context: context},
			stepName: stepName,
		},
		BaseEndPart: NewEndPart(usage, FinishReasonFailed),
		error:       err,
	}
}
