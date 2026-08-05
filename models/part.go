package models

type PartType string

const (
	PartTypeStart PartType = "start"
	PartTypeEnd   PartType = "end"
	PartTypeText  PartType = "text"

	PartTypeToolStart  PartType = "tool_start"
	PartTypeToolResult PartType = "tool_result"
	PartTypeToolError  PartType = "tool_error"

	PartTypeStepStart PartType = "step_start"
	PartTypeStepEnd   PartType = "step_end"
)

type Part interface {
	String() string
	GetType() PartType
}

type TextPart struct {
	Text string `json:"text,omitempty"`
}

func (t TextPart) String() string {
	return t.Text
}

func (t TextPart) GetType() PartType {
	return PartTypeText
}

type GeneralPart struct {
	Type PartType               `json:"type"`
	Data map[string]interface{} `json:"data,omitempty"`
}

func (g GeneralPart) String() string {
	if g.Type == PartTypeText {
		return g.Data["text"].(string)
	}
	return ""
}

func (g GeneralPart) GetType() PartType {
	return g.Type
}

func (e GeneralPart) ToSpecificEvent() Part {
	switch e.Type {
	case PartTypeText:
		return TextPart{
			Text: e.Data["text"].(string),
		}
	default:
		return e
	}
}
