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

type Part interface {
	GetContext() LanguageModelContext
	GetType() PartType
	GetModelName() string
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
