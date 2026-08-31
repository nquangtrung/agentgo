package models

import (
	"fmt"
	"sync"
)

type LanguageModelUsage struct {
	InputTokens         int64
	InputTokensDetails  LanguageModelUsageInputTokensDetails
	OutputTokens        int64
	OutputTokensDetails LanguageModelUsageOutputTokensDetails
	TotalTokens         int64
}

type LanguageModelUsageInputTokensDetails struct {
	CachedTokens     int64
	CacheWriteTokens int64
}

type LanguageModelUsageOutputTokensDetails struct {
	ReasoningTokens int64
}

func NewLanguageModelUsage(outputTokens, inputTokens, cachedTokens, reasoningTokens int64) LanguageModelUsage {
	return LanguageModelUsage{
		InputTokens: inputTokens,
		InputTokensDetails: LanguageModelUsageInputTokensDetails{
			CachedTokens: cachedTokens,
		},
		OutputTokens: outputTokens,
		OutputTokensDetails: LanguageModelUsageOutputTokensDetails{
			ReasoningTokens: reasoningTokens,
		},
		TotalTokens: inputTokens + outputTokens,
	}
}

type LanguageModelOutput struct {
	Text      string
	Usage     LanguageModelUsage
	ModelName string
	Context   *ToolExecutionsArchive
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

type ToolExecutionRecord struct {
	Name       string
	ToolCalled *ToolCall
	ToolResult *ToolExecuteOutput
}
type ToolExecutionsArchive struct {
	LanguageModelContext

	recordsLocker sync.Mutex
	records       []ToolExecutionRecord
}

type ContextKey string

const (
	ProviderContextKey      ContextKey = "provider_context"
	MachineContextKey       ContextKey = "machine_context"
	EndConditionsContextKey ContextKey = "end_conditions_context"
	ToolsContextKey         ContextKey = "tools_context"
	StreamContextKey        ContextKey = "stream_context"
	PartEmitterContextKey   ContextKey = "part_emitter_context"
	AccumulatorContextKey   ContextKey = "accumulator_context"
)

func (e *ToolExecutionsArchive) ModelName() string {
	return e.LanguageModelContext.ModelName
}

func (e *ToolExecutionsArchive) AddToolCall(stepName string, toolCalled *ToolCall) {
	e.recordsLocker.Lock()
	defer e.recordsLocker.Unlock()

	record := ToolExecutionRecord{
		Name:       stepName,
		ToolCalled: toolCalled,
	}
	e.records = append(e.records, record)
}

func (e *ToolExecutionsArchive) AddToolCallWithResult(stepName string, toolResult *ToolExecuteOutput) {
	e.recordsLocker.Lock()
	defer e.recordsLocker.Unlock()

	step := ToolExecutionRecord{
		Name:       stepName,
		ToolCalled: toolResult.ToolCall,
		ToolResult: toolResult,
	}

	e.records = append(e.records, step)
}

func (e *ToolExecutionsArchive) Records() []ToolExecutionRecord {
	e.recordsLocker.Lock()
	defer e.recordsLocker.Unlock()

	return e.records
}
func (e *ToolExecutionsArchive) RecordCount() int {
	e.recordsLocker.Lock()
	defer e.recordsLocker.Unlock()

	return len(e.records)
}

func (e *ToolExecutionsArchive) Record(index int) (*ToolExecutionRecord, error) {
	e.recordsLocker.Lock()
	defer e.recordsLocker.Unlock()

	if index < 0 || index >= len(e.records) {
		return nil, fmt.Errorf("step index out of range")
	}
	return &e.records[index], nil
}
func (e *ToolExecutionsArchive) LastRecord() *ToolExecutionRecord {
	e.recordsLocker.Lock()
	defer e.recordsLocker.Unlock()

	if len(e.records) == 0 {
		return nil
	}
	return &e.records[len(e.records)-1]
}
func (e *ToolExecutionsArchive) UpdateLastRecordResult(result *ToolExecuteOutput) {
	e.recordsLocker.Lock()
	defer e.recordsLocker.Unlock()

	if len(e.records) == 0 {
		return
	}

	result.ToolCall = e.records[len(e.records)-1].ToolCalled
	e.records[len(e.records)-1].ToolResult = result
}
func (e *ToolExecutionsArchive) UpdateLastRecordError(err error) {
	e.UpdateLastRecordResult(&ToolExecuteOutput{
		Error: err,
	})
}

//go:generate mockgen -source=model.go -destination=../tests/mocks/mock_context.go -package=mocks
func NewExecutionContextFromLanguageModelContext(context LanguageModelContext) *ToolExecutionsArchive {
	return &ToolExecutionsArchive{
		LanguageModelContext: context,
		records:              []ToolExecutionRecord{},
	}
}
