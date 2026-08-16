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
	ModelName() string
	StepName() string
	StepIndex() int
	NextStep(stepName string) ExecutionContext
}
type BaseExecutionContext struct {
	LanguageModelContext LanguageModelContext

	stepName  string
	stepIndex int
}

func (e BaseExecutionContext) ModelName() string {
	return e.LanguageModelContext.ModelName
}

func (e BaseExecutionContext) StepName() string {
	return e.stepName
}

func (e BaseExecutionContext) StepIndex() int {
	return e.stepIndex
}

func (e BaseExecutionContext) NextStep(stepName string) ExecutionContext {
	return BaseExecutionContext{
		LanguageModelContext: e.LanguageModelContext,
		stepName:             stepName,
		stepIndex:            e.stepIndex + 1,
	}
}

func ExecutionContextFromLanguageModelContext(context LanguageModelContext) ExecutionContext {
	return BaseExecutionContext{
		LanguageModelContext: context,
		stepName:             "",
		stepIndex:            0,
	}
}
