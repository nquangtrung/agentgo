package models

type ContentPart interface {
	Text() string
}

type TextPart interface {
	Part
	ContentPart
}

type BaseTextPart struct {
	BasePart
	text string
}

func (t BaseTextPart) Text() string {
	return t.text
}

func NewTextPart(context LanguageModelContext, text string) *BaseTextPart {
	return &BaseTextPart{
		BasePart: BasePart{
			partType: PartTypeText,
			context:  context,
		},
		text: text,
	}
}
