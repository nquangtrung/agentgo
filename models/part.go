package models

import (
	"time"
)

type PartType string

const (
	PartTypeStart PartType = "start"
	PartTypeEnd   PartType = "end"
	PartTypeText  PartType = "text"

	PartTypeToolStart  PartType = "tool_start"
	PartTypeToolResult PartType = "tool_result"
	PartTypeToolError  PartType = "tool_error"

	PartTypeStepStart PartType = "step_start"
	PartTypeStepEnd   PartType = "step_end"
	PartTypeStepError PartType = "step_error"
)

type FinishReason string

const (
	FinishReasonCompleted FinishReason = "completed"
	FinishReasonFailed    FinishReason = "failed"
	FinishReasonCancelled FinishReason = "cancelled"
	FinishReasonTimeout   FinishReason = "timeout"
)

type Part interface {
	Context() LanguageModelContext
	Type() PartType
	ModelName() string
	isPart()
}

type BasePart struct {
	partType PartType
	context  LanguageModelContext
}

func (p BasePart) Type() PartType                { return p.partType }
func (p BasePart) ModelName() string             { return p.context.ModelName }
func (p BasePart) Context() LanguageModelContext { return p.context }
func (p BasePart) isPart()                       {}

func NewPart(context LanguageModelContext, partType PartType) BasePart {
	return BasePart{partType: partType, context: context}
}

type StartPart interface {
	At() time.Time
}

type BaseStartPart struct {
	at time.Time
}

func (s *BaseStartPart) At() time.Time { return s.at }

type EndPart interface {
	Usage() LanguageModelUsage
	FinishReason() FinishReason
}

type BaseEndPart struct {
	usage        LanguageModelUsage
	finishReason FinishReason
}

func (p BaseEndPart) Usage() LanguageModelUsage  { return p.usage }
func (p BaseEndPart) FinishReason() FinishReason { return p.finishReason }

func NewEndPart(usage LanguageModelUsage, finishReason FinishReason) BaseEndPart {
	return BaseEndPart{usage: usage, finishReason: finishReason}
}

type ProcessStartPart interface {
	Part
	StartPart
}

type BaseProcessStartPart struct {
	BasePart
	BaseStartPart
}

func NewProcessStartPart(context LanguageModelContext) *BaseProcessStartPart {
	return &BaseProcessStartPart{
		BasePart:      BasePart{partType: PartTypeStart, context: context},
		BaseStartPart: BaseStartPart{at: time.Now()},
	}
}

type ProcessEndPart interface {
	Part
	EndPart
}

type BaseProcessEndPart struct {
	BasePart
	BaseEndPart
}

func NewProcessEndPart(context LanguageModelContext, usage LanguageModelUsage, finishReason FinishReason) *BaseProcessEndPart {
	return &BaseProcessEndPart{
		BasePart:    BasePart{partType: PartTypeEnd, context: context},
		BaseEndPart: NewEndPart(usage, finishReason),
	}
}
