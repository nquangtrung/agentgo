package models

type LanguageModelUsage struct {
	OutputTokens    int
	InputTokens     int
	CachedTokens    int
	ReasoningTokens int
}

func NewLanguageModelUsage(outputTokens, inputTokens, cachedTokens, reasoningTokens int) LanguageModelUsage {
	return LanguageModelUsage{
		OutputTokens:    outputTokens,
		InputTokens:     inputTokens,
		CachedTokens:    cachedTokens,
		ReasoningTokens: reasoningTokens,
	}
}

type LanguageModelOutput struct {
	Text      string
	Usage     LanguageModelUsage
	ModelName string
}

func NewLanguageModelOutput(text string, usage LanguageModelUsage, modelName string) LanguageModelOutput {
	return LanguageModelOutput{
		Text:      text,
		Usage:     usage,
		ModelName: modelName,
	}
}

type LanguageModelStreamOutput struct {
	Channel   chan Part
	ModelName string
}

func NewLanguageModelStreamOutput(channel chan Part, modelName string) LanguageModelStreamOutput {
	return LanguageModelStreamOutput{
		Channel:   channel,
		ModelName: modelName,
	}
}

type LanguageModelContext struct {
	ModelName string
}

func NewLanguageModelContext(modelName string) LanguageModelContext {
	return LanguageModelContext{
		ModelName: modelName,
	}
}

type ExecutionContext interface {
	GetModelName() string
	GetStepName() string
	GetStepIndex() int
	NextStep(stepName string) ExecutionContext
}
type ExecutionContextImpl struct {
	LanguageModelContext LanguageModelContext

	stepName  string
	stepIndex int
}

func (e ExecutionContextImpl) GetModelName() string {
	return e.LanguageModelContext.ModelName
}

func (e ExecutionContextImpl) GetStepName() string {
	return e.stepName
}

func (e ExecutionContextImpl) GetStepIndex() int {
	return e.stepIndex
}

func (e ExecutionContextImpl) NextStep(stepName string) ExecutionContext {
	return ExecutionContextImpl{
		LanguageModelContext: e.LanguageModelContext,
		stepName:             stepName,
		stepIndex:            e.stepIndex + 1,
	}
}

func ExecutionContextFromLanguageModelContext(context LanguageModelContext) ExecutionContext {
	return ExecutionContextImpl{
		LanguageModelContext: context,
		stepName:             "",
		stepIndex:            0,
	}
}
