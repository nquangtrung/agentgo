package models

type ToolExecuteParams struct {
	Params map[string]any
}

type ToolExecuteOutput struct {
	Result map[string]any
	Error  error
	Usage  LanguageModelUsage
}

type ToolCall struct {
	ToolName string
	Params   map[string]any
}

//go:generate mockgen -destination=../mocks/mock_tool.go -package=mocks trontria.com/agentgo/models Tool
type Tool interface {
	Name() string
	Description() string
	Execute(params ToolExecuteParams) ToolExecuteOutput
}

type BaseTool struct {
	name         string
	description  string
	parameters   map[string]any
	fn           func(params ToolExecuteParams) ToolExecuteOutput
	outputSchema map[string]any
	strict       bool
}

func (t BaseTool) Name() string {
	return t.name
}

func (t BaseTool) Description() string {
	return t.description
}

func (t BaseTool) Execute(params ToolExecuteParams) ToolExecuteOutput {
	return t.fn(params)
}
