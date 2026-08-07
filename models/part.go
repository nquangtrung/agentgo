package models

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

type AsPart interface {
	AsTextPart() (TextPart, bool)
	AsStepStartPart() (StepStartPart, bool)
	AsStepEndPart() (StepEndPart, bool)
	AsToolStartPart() (ToolPart, bool)
	AsToolResultPart() (ToolResultPart, bool)
	AsToolErrorPart() (ToolErrorPart, bool)
}

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

func (p PartImpl) AsToolStartPart() (ToolPart, bool) {
	if toolStartPart, ok := interface{}(p).(ToolPart); ok {
		return toolStartPart, true
	}
	return nil, false
}

func (p PartImpl) AsToolResultPart() (ToolResultPart, bool) {
	if toolResultPart, ok := interface{}(p).(ToolResultPart); ok {
		return toolResultPart, true
	}
	return nil, false
}

func (p PartImpl) AsToolErrorPart() (ToolErrorPart, bool) {
	if toolErrorPart, ok := interface{}(p).(ToolErrorPart); ok {
		return toolErrorPart, true
	}
	return nil, false
}

func (p PartImpl) AsTextPart() (TextPart, bool) {
	if textPart, ok := interface{}(p).(TextPart); ok {
		return textPart, true
	}
	return nil, false
}

func (p PartImpl) AsStepStartPart() (StepStartPart, bool) {
	if stepStartPart, ok := interface{}(p).(StepStartPart); ok {
		return stepStartPart, true
	}
	return nil, false
}

func (p PartImpl) AsStepEndPart() (StepEndPart, bool) {
	if stepEndPart, ok := interface{}(p).(StepEndPart); ok {
		return stepEndPart, true
	}
	return nil, false
}

func NewPart(context LanguageModelContext, partType PartType) PartImpl {
	return PartImpl{
		partType: partType,
		context:  context,
	}
}

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
