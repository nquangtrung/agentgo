package models

type ToolPart interface {
	GetToolName() string
}

type ToolResultPart interface {
	Part
	ToolPart
	EndPart
	GetResult() map[string]any
	GetUsage() LanguageModelUsage
}

type ToolErrorPart interface {
	Part
	ToolPart
	EndPart
	GetError() map[string]any
}

type ToolPartImpl struct {
	PartImpl
	toolName string
}

func (t ToolPartImpl) GetToolName() string {
	return t.toolName
}

type ToolStartPart interface {
	Part
	ToolPart
}

type ToolStartPartImpl struct {
	ToolPartImpl
	data map[string]any
}

type ToolResultPartImpl struct {
	ToolPartImpl
	EndPartImpl
	data map[string]any
}

func (t ToolResultPartImpl) GetResult() map[string]any {
	return t.data
}

type ToolErrorPartImpl struct {
	ToolPartImpl
	EndPartImpl
	error map[string]any
}

func (t ToolErrorPartImpl) GetError() map[string]any {
	return t.error
}

func NewToolStartPart(context LanguageModelContext, toolName string) ToolStartPart {
	return &ToolStartPartImpl{
		ToolPartImpl: ToolPartImpl{
			PartImpl: PartImpl{
				partType: PartTypeToolStart,
				context:  context,
			},
			toolName: toolName,
		},
	}
}

func NewToolResultPart(context LanguageModelContext, toolName string, result map[string]any, usage LanguageModelUsage) ToolResultPart {
	return &ToolResultPartImpl{
		EndPartImpl: NewEndPart(usage, FinishReasonCompleted),
		ToolPartImpl: ToolPartImpl{
			PartImpl: PartImpl{
				partType: PartTypeToolResult,
				context:  context,
			},
			toolName: toolName,
		},
		data: result,
	}
}

func NewToolErrorPart(context LanguageModelContext, toolName string, errorData map[string]any) ToolErrorPart {
	return &ToolErrorPartImpl{
		EndPartImpl: NewEndPart(LanguageModelUsage{}, FinishReasonFailed),
		ToolPartImpl: ToolPartImpl{
			PartImpl: PartImpl{
				partType: PartTypeToolError,
				context:  context,
			},
			toolName: toolName,
		},
		error: errorData,
	}
}
