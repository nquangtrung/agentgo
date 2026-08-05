package models

type ToolStartPart struct {
	ToolName string                 `json:"tool_name,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

func (t ToolStartPart) String() string {
	return t.ToolName
}

func (t ToolStartPart) GetType() PartType {
	return PartTypeToolStart
}

type ToolResultPart struct {
	ToolName string                 `json:"tool_name,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

func (t ToolResultPart) String() string {
	return t.ToolName
}

func (t ToolResultPart) GetType() PartType {
	return PartTypeToolResult
}

type ToolErrorPart struct {
	ToolName string                 `json:"tool_name,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}
