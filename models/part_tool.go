package models

type ToolPart interface {
	ToolName() string
}

type ToolResultPart interface {
	Part
	ToolPart
	EndPart
	Result() map[string]any
	Usage() LanguageModelUsage
}

type ToolErrorPart interface {
	Part
	ToolPart
	EndPart
	Error() map[string]any
}

type BaseToolPart struct {
	BasePart
	toolName string
}

func (t BaseToolPart) ToolName() string {
	return t.toolName
}

type ToolStartPart interface {
	Part
	ToolPart
}

type BaseToolStartPart struct {
	BaseToolPart
	data map[string]any
}

type BaseToolResultPart struct {
	BaseToolPart
	EndPartImpl
	data map[string]any
}

func (t BaseToolResultPart) Result() map[string]any {
	return t.data
}

type BaseToolErrorPart struct {
	BaseToolPart
	EndPartImpl
	error map[string]any
}

func (t BaseToolErrorPart) Error() map[string]any {
	return t.error
}

func NewToolStartPart(context LanguageModelContext, toolName string) *BaseToolStartPart {
	return &BaseToolStartPart{
		BaseToolPart: BaseToolPart{
			BasePart: BasePart{
				partType: PartTypeToolStart,
				context:  context,
			},
			toolName: toolName,
		},
	}
}

func NewToolResultPart(context LanguageModelContext, toolName string, result map[string]any, usage LanguageModelUsage) *BaseToolResultPart {
	return &BaseToolResultPart{
		EndPartImpl: NewEndPart(usage, FinishReasonCompleted),
		BaseToolPart: BaseToolPart{
			BasePart: BasePart{
				partType: PartTypeToolResult,
				context:  context,
			},
			toolName: toolName,
		},
		data: result,
	}
}

func NewToolErrorPart(context LanguageModelContext, toolName string, errorData map[string]any) *BaseToolErrorPart {
	return &BaseToolErrorPart{
		EndPartImpl: NewEndPart(LanguageModelUsage{}, FinishReasonFailed),
		BaseToolPart: BaseToolPart{
			BasePart: BasePart{
				partType: PartTypeToolError,
				context:  context,
			},
			toolName: toolName,
		},
		error: errorData,
	}
}
