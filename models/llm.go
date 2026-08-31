package models

type LanguageModelUsage struct {
	InputTokens         int64
	InputTokensDetails  LanguageModelUsageInputTokensDetails
	OutputTokens        int64
	OutputTokensDetails LanguageModelUsageOutputTokensDetails
	TotalTokens         int64
}

type LanguageModelUsageInputTokensDetails struct {
	CachedTokens     int64
	CacheWriteTokens int64
}

type LanguageModelUsageOutputTokensDetails struct {
	ReasoningTokens int64
}

type LanguageModelStreamOutput struct {
	Channel   chan Part
	ModelName string
}

func NewLanguageModelStreamOutput(channel chan Part, modelName string) LanguageModelStreamOutput {
	return LanguageModelStreamOutput{
		Channel:   channel,
		ModelName: modelName,
	}
}
func NewLanguageModelUsage(outputTokens, inputTokens, cachedTokens, reasoningTokens int64) LanguageModelUsage {
	return LanguageModelUsage{
		InputTokens: inputTokens,
		InputTokensDetails: LanguageModelUsageInputTokensDetails{
			CachedTokens: cachedTokens,
		},
		OutputTokens: outputTokens,
		OutputTokensDetails: LanguageModelUsageOutputTokensDetails{
			ReasoningTokens: reasoningTokens,
		},
		TotalTokens: inputTokens + outputTokens,
	}
}

type LanguageModelContext struct {
	ModelName string
}

func NewLanguageModelContext(modelName string) LanguageModelContext {
	return LanguageModelContext{ModelName: modelName}
}
