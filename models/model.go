package models

import (
	"fmt"
	"sync"
)

type LanguageModelUsage struct {
	OutputTokens    int64
	InputTokens     int64
	CachedTokens    int64
	ReasoningTokens int64
}

func NewLanguageModelUsage(outputTokens, inputTokens, cachedTokens, reasoningTokens int64) LanguageModelUsage {
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

type ContextKey string

const (
	ProviderContextKey      ContextKey = "provider_context"
	RunnerContextKey        ContextKey = "runner_context"
	MachineContextKey       ContextKey = "machine_context"
	EndConditionsContextKey ContextKey = "end_conditions_context"
	ToolsContextKey         ContextKey = "tools_context"
)

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

func (e *ExecutionContext) AddStepWithResult(stepName string, toolResult *ToolExecuteOutput) {
	e.stepLocker.Lock()
	defer e.stepLocker.Unlock()

	step := Step{
		Name:       stepName,
		ToolCalled: toolResult.ToolCall,
		ToolResult: toolResult,
	}

	e.steps = append(e.steps, step)
}

func (e *ExecutionContext) Steps() []Step {
	e.stepLocker.Lock()
	defer e.stepLocker.Unlock()

	return e.steps
}
func (e *ExecutionContext) StepCount() int {
	e.stepLocker.Lock()
	defer e.stepLocker.Unlock()

	return len(e.steps)
}

func (e *ExecutionContext) Step(index int) (*Step, error) {
	e.stepLocker.Lock()
	defer e.stepLocker.Unlock()

	if index < 0 || index >= len(e.steps) {
		return nil, fmt.Errorf("step index out of range")
	}
	return &e.steps[index], nil
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

	result.ToolCall = e.steps[len(e.steps)-1].ToolCalled
	e.steps[len(e.steps)-1].ToolResult = result
}
func (e *ExecutionContext) UpdateLastStepError(err error) {
	e.UpdateLastStepResult(&ToolExecuteOutput{
		Error: err,
	})
}

//go:generate mockgen -source=model.go -destination=../tests/mocks/mock_context.go -package=mocks
func NewExecutionContextFromLanguageModelContext(context LanguageModelContext) *ExecutionContext {
	return &ExecutionContext{
		LanguageModelContext: context,
		steps:                []Step{},
	}
}
