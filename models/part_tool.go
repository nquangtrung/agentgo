package models

type ToolPart interface {
	Part
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
	ToolName string `json:"tool_name,omitempty"`
}

func (t ToolPartImpl) GetToolName() string {
	return t.ToolName
}

type ToolStartPartImpl struct {
	ToolPartImpl
	Data map[string]interface{} `json:"data,omitempty"`
}

type ToolResultPartImpl struct {
	ToolPartImpl
	EndPartImpl
	Data map[string]interface{} `json:"data,omitempty"`
}

func (t ToolResultPartImpl) GetResult() map[string]interface{} {
	return t.Data
}

type ToolErrorPartImpl struct {
	ToolPartImpl
	EndPartImpl
	Error map[string]interface{} `json:"error,omitempty"`
}

func (t ToolErrorPartImpl) GetError() map[string]interface{} {
	return t.Error
}

func NewToolStartPart(context LanguageModelContext, toolName string) ToolPart {
	return ToolStartPartImpl{
		ToolPartImpl: ToolPartImpl{
			PartImpl: NewPart(context, PartTypeToolStart),
			ToolName: toolName,
		},
	}
}

func NewToolResultPart(context LanguageModelContext, toolName string, result map[string]interface{}, usage LanguageModelUsage) ToolResultPart {
	return ToolResultPartImpl{
		EndPartImpl: NewPartEnd(usage, FinishReasonCompleted),
		ToolPartImpl: ToolPartImpl{
			PartImpl: NewPart(context, PartTypeToolResult),
			ToolName: toolName,
		},
		Data: result,
	}
}

func NewToolErrorPart(context LanguageModelContext, toolName string, errorData map[string]interface{}) ToolErrorPart {
	return ToolErrorPartImpl{
		EndPartImpl: NewPartEnd(LanguageModelUsage{}, FinishReasonFailed),
		ToolPartImpl: ToolPartImpl{
			PartImpl: NewPart(context, PartTypeToolError),
			ToolName: toolName,
		},
		Error: errorData,
	}
}
