package models

type ToolPart interface {
	GetToolName() string
}

type ToolResultPart interface {
	ToolPart
	EndPart
	GetResult() map[string]interface{}
	GetUsage() LanguageModelUsage
}

type ToolErrorPart interface {
	ToolPart
	EndPart
	GetError() map[string]interface{}
}

type ToolPartImpl struct {
	PartImpl
	toolName string
}

func (t ToolPartImpl) GetToolName() string {
	return t.toolName
}

type ToolStartPartImpl struct {
	ToolPartImpl
	data map[string]interface{}
}

type ToolResultPartImpl struct {
	ToolPartImpl
	EndPartImpl
	data map[string]interface{}
}

func (t ToolResultPartImpl) GetResult() map[string]interface{} {
	return t.data
}

type ToolErrorPartImpl struct {
	ToolPartImpl
	EndPartImpl
	error map[string]interface{}
}

func (t ToolErrorPartImpl) GetError() map[string]interface{} {
	return t.error
}

func NewToolStartPart(context LanguageModelContext, toolName string) ToolPart {
	return ToolStartPartImpl{
		ToolPartImpl: ToolPartImpl{
			PartImpl: NewPart(context, PartTypeToolStart),
			toolName: toolName,
		},
	}
}

func NewToolResultPart(context LanguageModelContext, toolName string, result map[string]interface{}, usage LanguageModelUsage) ToolResultPart {
	return ToolResultPartImpl{
		EndPartImpl: NewEndPart(usage, FinishReasonCompleted),
		ToolPartImpl: ToolPartImpl{
			PartImpl: NewPart(context, PartTypeToolResult),
			toolName: toolName,
		},
		data: result,
	}
}

func NewToolErrorPart(context LanguageModelContext, toolName string, errorData map[string]interface{}) ToolErrorPart {
	return ToolErrorPartImpl{
		EndPartImpl: NewEndPart(LanguageModelUsage{}, FinishReasonFailed),
		ToolPartImpl: ToolPartImpl{
			PartImpl: NewPart(context, PartTypeToolError),
			toolName: toolName,
		},
		error: errorData,
	}
}
