package models

import "time"

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

//go:generate mockgen -destination=../mocks/mock_part.go -package=mocks trontria.com/agentgo/models Part
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
	return BasePart{
		partType: partType,
		context:  context,
	}
}

//go:generate mockgen -destination=../mocks/mock_start_part.go -package=mocks trontria.com/agentgo/models StartPart
type StartPart interface {
	At() time.Time
}
type StartPartImpl struct {
	at time.Time
}

func (s *StartPartImpl) At() time.Time { return s.at }

//go:generate mockgen -destination=../mocks/mock_end_part.go -package=mocks trontria.com/agentgo/models EndPart
type EndPart interface {
	Usage() LanguageModelUsage
	FinishReason() FinishReason
}

type EndPartImpl struct {
	usage        LanguageModelUsage
	finishReason FinishReason
}

func (p EndPartImpl) Usage() LanguageModelUsage  { return p.usage }
func (p EndPartImpl) FinishReason() FinishReason { return p.finishReason }

func NewEndPart(usage LanguageModelUsage, finishReason FinishReason) EndPartImpl {
	return EndPartImpl{
		usage:        usage,
		finishReason: finishReason,
	}
}
