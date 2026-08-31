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
	Error() error
}

type BaseToolPart struct {
	BasePart
	toolName string
}

func (t BaseToolPart) ToolName() string { return t.toolName }

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
	BaseEndPart
	data map[string]any
}

func (t BaseToolResultPart) Result() map[string]any { return t.data }

type BaseToolErrorPart struct {
	BaseToolPart
	BaseEndPart
	error error
}

func (t BaseToolErrorPart) Error() error { return t.error }

func NewToolStartPart(context LanguageModelContext, toolName string) *BaseToolStartPart {
	return &BaseToolStartPart{
		BaseToolPart: BaseToolPart{
			BasePart: BasePart{partType: PartTypeToolStart, context: context},
			toolName: toolName,
		},
	}
}

func NewToolResultPart(context LanguageModelContext, toolName string, result ToolExecuteOutput) *BaseToolResultPart {
	return &BaseToolResultPart{
		BaseEndPart: NewEndPart(llmUsageFromModels(result.Usage), FinishReasonCompleted),
		BaseToolPart: BaseToolPart{
			BasePart: BasePart{partType: PartTypeToolResult, context: context},
			toolName: toolName,
		},
		data: result.Output,
	}
}

func NewToolErrorPart(context LanguageModelContext, toolName string, usage LanguageModelUsage, err error) *BaseToolErrorPart {
	return &BaseToolErrorPart{
		BaseEndPart: NewEndPart(usage, FinishReasonFailed),
		BaseToolPart: BaseToolPart{
			BasePart: BasePart{partType: PartTypeToolError, context: context},
			toolName: toolName,
		},
		error: err,
	}
}

func llmUsageFromModels(u LanguageModelUsage) LanguageModelUsage {
	return LanguageModelUsage{
		InputTokens: u.InputTokens,
		InputTokensDetails: LanguageModelUsageInputTokensDetails{
			CachedTokens:     u.InputTokensDetails.CachedTokens,
			CacheWriteTokens: u.InputTokensDetails.CacheWriteTokens,
		},
		OutputTokens: u.OutputTokens,
		OutputTokensDetails: LanguageModelUsageOutputTokensDetails{
			ReasoningTokens: u.OutputTokensDetails.ReasoningTokens,
		},
		TotalTokens: u.TotalTokens,
	}
}
