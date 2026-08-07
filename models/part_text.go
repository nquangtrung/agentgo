package models

type ContentPart interface {
	GetText() string
}

type TextPart interface {
	Part
	ContentPart
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
