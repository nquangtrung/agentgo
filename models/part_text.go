package models

type TextPart interface {
	Part
	GetText() string
}

type TextPartImpl struct {
	PartImpl
	text string
}

func (t TextPartImpl) GetText() string {
	return t.text
}

func NewTextPart(context LanguageModelContext, text string) TextPart {
	return TextPartImpl{
		PartImpl: NewPart(context, PartTypeText),
		text:     text,
	}
}

func ToTextPart(part Part) (TextPart, bool) {
	if textPart, ok := part.(TextPart); ok {
		return textPart, true
	}
	return nil, false
}
