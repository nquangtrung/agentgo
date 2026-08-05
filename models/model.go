package models

type LanguageModelUsage struct {
	OutputTokens    int
	InputTokens     int
	CachedTokens    int
	ReasoningTokens int
}

type LanguageModelOutput struct {
	Text  string
	Usage LanguageModelUsage
	Model string
}
