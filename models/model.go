package models

import (
	"sync"
)

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
	Context   *ExecutionContext
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

type Step struct {
	Name       string
	ToolCalled *ToolCall
	ToolResult *ToolExecuteOutput
}
type ExecutionContext struct {
	LanguageModelContext

	stepLocker sync.Mutex
	steps      []Step
}

func (e *ExecutionContext) ModelName() string {
	return e.LanguageModelContext.ModelName
}

func (e *ExecutionContext) AddStep(stepName string, toolCalled *ToolCall) {
	e.stepLocker.Lock()
	defer e.stepLocker.Unlock()

	step := Step{
		Name:       stepName,
		ToolCalled: toolCalled,
	}
	e.steps = append(e.steps, step)
}
func (e *ExecutionContext) Steps() []Step {
	e.stepLocker.Lock()
	defer e.stepLocker.Unlock()

	return e.steps
}
func (e *ExecutionContext) LastStep() *Step {
	e.stepLocker.Lock()
	defer e.stepLocker.Unlock()

	if len(e.steps) == 0 {
		return nil
	}
	return &e.steps[len(e.steps)-1]
}
func (e *ExecutionContext) UpdateLastStepResult(result *ToolExecuteOutput) {
	e.stepLocker.Lock()
	defer e.stepLocker.Unlock()

	if len(e.steps) == 0 {
		return
	}
	e.steps[len(e.steps)-1].ToolResult = result
}
func (e *ExecutionContext) UpdateLastStepError(err error) {
	e.UpdateLastStepResult(&ToolExecuteOutput{
		Error: err,
	})
}

func NewExecutionContextFromLanguageModelContext(context LanguageModelContext) ExecutionContext {
	return ExecutionContext{
		LanguageModelContext: context,
		steps:                []Step{},
	}
}
