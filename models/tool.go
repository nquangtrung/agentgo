package models

type ToolExecuteParams struct {
	Params map[string]interface{}
}

type ToolExecuteOutput struct {
	Result map[string]interface{}
	Error  map[string]interface{}
}

type ToolCall struct {
	ToolName string
	Params   map[string]interface{}
}

type Tool interface {
	GetName() string
	GetDescription() string
	Execute(params ToolExecuteParams) (ToolExecuteOutput, error)
}

type ToolParams struct {
	Name         string
	Description  string
	Parameters   map[string]interface{}
	Fn           func(params ToolExecuteParams) (ToolExecuteOutput, error)
	OutputSchema map[string]interface{}
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
