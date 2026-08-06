package models

type TextPart interface {
	Part
	GetText() string
}

type TextPartImpl struct {
	PartImpl
	Text string `json:"text,omitempty"`
}

func (t TextPartImpl) GetText() string {
	return t.Text
}

func NewTextPart(context LanguageModelContext, text string) TextPart {
	return TextPartImpl{
		PartImpl: NewPart(context, PartTypeText),
		Text:     text,
	}
}

func ToTextPart(part Part) (TextPart, bool) {
	if textPart, ok := part.(TextPart); ok {
		return textPart, true
	}
	return nil, false
}
