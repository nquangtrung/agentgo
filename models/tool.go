package models

type ToolExecuteParams struct {
	Params map[string]any
}

type ToolExecuteOutput struct {
	Result   map[string]any
	Error    error
	Usage    LanguageModelUsage
	ToolCall *ToolCall
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
	name        string
	description string
	strict      bool

	fn func(params ToolExecuteParams) ToolExecuteOutput

	inputSchema  map[string]any
	outputSchema map[string]any
}

func (t BaseTool) InputSchema() map[string]any {
	return t.inputSchema
}

func (t BaseTool) OutputSchema() map[string]any {
	return t.outputSchema
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

type NewToolParams struct {
	Name        string
	Description string
	Fn          func(params ToolExecuteParams) ToolExecuteOutput

	InputSchema  map[string]any
	OutputSchema map[string]any
}

func NewTool(params NewToolParams) BaseTool {
	return BaseTool{
		name:         params.Name,
		description:  params.Description,
		fn:           params.Fn,
		inputSchema:  params.InputSchema,
		outputSchema: params.OutputSchema,
	}
}
