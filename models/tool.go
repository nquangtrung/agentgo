package models

type ToolExecuteParams struct {
	Params map[string]any
}

type ToolExecuteOutput struct {
	Result map[string]any
	Error  map[string]any
}

type ToolCall struct {
	ToolName string
	Params   map[string]any
}

//go:generate mockgen -destination=../mocks/mock_tool.go -package=mocks trontria.com/agentgo/models Tool
type Tool interface {
	GetName() string
	GetDescription() string
	Execute(params ToolExecuteParams) (ToolExecuteOutput, error)
}

type ToolParams struct {
	Name         string
	Description  string
	Parameters   map[string]any
	Fn           func(params ToolExecuteParams) (ToolExecuteOutput, error)
	OutputSchema map[string]any
	Strict       bool
}

func (t ToolParams) GetName() string {
	return t.Name
}

func (t ToolParams) GetDescription() string {
	return t.Description
}

func (t ToolParams) Execute(params ToolExecuteParams) (ToolExecuteOutput, error) {
	return t.Fn(params)
}
