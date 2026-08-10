package models

import "unsafe"

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
)

type FinishReason string

const (
	FinishReasonCompleted FinishReason = "completed"
	FinishReasonFailed    FinishReason = "failed"
	FinishReasonCancelled FinishReason = "cancelled"
	FinishReasonTimeout   FinishReason = "timeout"
)

//go:generate mockgen -destination=../mocks/mock_as_part.go -package=mocks trontria.com/agentgo/models AsPart
type AsPart interface {
	AsTextPart() (TextPart, bool)
	AsStepStartPart() (StepStartPart, bool)
	AsStepEndPart() (StepEndPart, bool)
	AsToolStartPart() (ToolPart, bool)
	AsToolResultPart() (ToolResultPart, bool)
	AsToolErrorPart() (ToolErrorPart, bool)
}

//go:generate mockgen -destination=../mocks/mock_part.go -package=mocks trontria.com/agentgo/models Part
type Part interface {
	GetContext() LanguageModelContext
	GetType() PartType
	GetModelName() string

	AsPart
}

type PartImpl struct {
	partType PartType
	context  LanguageModelContext
}

func (p PartImpl) GetType() PartType {
	return p.partType
}

func (p PartImpl) GetModelName() string {
	return p.context.ModelName
}

func (p PartImpl) GetContext() LanguageModelContext {
	return p.context
}

func (p *PartImpl) AsToolStartPart() (ToolPart, bool) {
	if p.partType == PartTypeToolStart {
		// Use unsafe pointer to reconstruct the full concrete type
		fullPtr := (*ToolStartPartImpl)(unsafe.Pointer(p))
		if toolStartPart, ok := any(fullPtr).(ToolPart); ok {
			return toolStartPart, true
		}
	}
	return nil, false
}

func (p *PartImpl) AsToolResultPart() (ToolResultPart, bool) {
	if p.partType == PartTypeToolResult {
		// Use unsafe pointer to reconstruct the full concrete type
		fullPtr := (*ToolResultPartImpl)(unsafe.Pointer(p))
		if toolResultPart, ok := any(fullPtr).(ToolResultPart); ok {
			return toolResultPart, true
		}
	}
	return nil, false
}

func (p *PartImpl) AsToolErrorPart() (ToolErrorPart, bool) {
	if p.partType == PartTypeToolError {
		// Use unsafe pointer to reconstruct the full concrete type
		fullPtr := (*ToolErrorPartImpl)(unsafe.Pointer(p))
		if toolErrorPart, ok := any(fullPtr).(ToolErrorPart); ok {
			return toolErrorPart, true
		}
	}
	return nil, false
}

func (p *PartImpl) AsTextPart() (TextPart, bool) {
	if p.partType == PartTypeText {
		// Use unsafe pointer to reconstruct the full concrete type
		fullPtr := (*TextPartImpl)(unsafe.Pointer(p))
		if textPart, ok := any(fullPtr).(TextPart); ok {
			return textPart, true
		}
	}
	return nil, false
}

func (p *PartImpl) AsStepStartPart() (StepStartPart, bool) {
	if p.partType == PartTypeStepStart {
		// Use unsafe pointer to reconstruct the full concrete type
		fullPtr := (*StepStartPartImpl)(unsafe.Pointer(p))
		if stepStartPart, ok := any(fullPtr).(StepStartPart); ok {
			return stepStartPart, true
		}
	}
	return nil, false
}

func (p *PartImpl) AsStepEndPart() (StepEndPart, bool) {
	if p.partType == PartTypeStepEnd {
		// Use unsafe pointer to reconstruct the full concrete type
		fullPtr := (*StepEndPartImpl)(unsafe.Pointer(p))
		if stepEndPart, ok := any(fullPtr).(StepEndPart); ok {
			return stepEndPart, true
		}
	}
	return nil, false
}

func NewPart(context LanguageModelContext, partType PartType) PartImpl {
	return PartImpl{
		partType: partType,
		context:  context,
	}
}

//go:generate mockgen -destination=../mocks/mock_end_part.go -package=mocks trontria.com/agentgo/models EndPart
type EndPart interface {
	GetUsage() LanguageModelUsage
	GetFinishReason() FinishReason
}

type EndPartImpl struct {
	Usage        LanguageModelUsage `json:"usage,omitempty"`
	FinishReason FinishReason       `json:"finish_reason,omitempty"`
}

func (p EndPartImpl) GetUsage() LanguageModelUsage {
	return p.Usage
}

func (p EndPartImpl) GetFinishReason() FinishReason {
	return p.FinishReason
}

func NewEndPart(usage LanguageModelUsage, finishReason FinishReason) EndPartImpl {
	return EndPartImpl{
		Usage:        usage,
		FinishReason: finishReason,
	}
}
