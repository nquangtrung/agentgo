package models

import (
	"log"
)

func AccumulateToolCallResult(acc *ToolExecutionsArchive, item *ToolExecuteOutput, messages *[]Message) {
	log.Printf("Accumulator called with item: %+v", item)
	switch {
	case item == nil:
		log.Printf("Accumulator received nil item, skipping")
		return
	case item.ToolCall != nil:
		log.Printf("Adding tool call result to execution context: %s", item.ToolCall.ToolName)
		acc.AddToolCallWithResult(item.ToolCall.ToolName, item)
		*messages = append(*messages, NewMessageFromToolResult(*item))
		return
	default:
		log.Printf("Adding text output to execution context: %s", item.Output)
		acc.AddToolCallWithResult("text", item)
	}
}

func AccumulateUsage(u1 LanguageModelUsage, u2 LanguageModelUsage) LanguageModelUsage {
	return LanguageModelUsage{
		InputTokens: u1.InputTokens + u2.InputTokens,
		InputTokensDetails: LanguageModelUsageInputTokensDetails{
			CachedTokens:     u1.InputTokensDetails.CachedTokens + u2.InputTokensDetails.CachedTokens,
			CacheWriteTokens: u1.InputTokensDetails.CacheWriteTokens + u2.InputTokensDetails.CacheWriteTokens,
		},
		OutputTokens: u1.OutputTokens + u2.OutputTokens,
		OutputTokensDetails: LanguageModelUsageOutputTokensDetails{
			ReasoningTokens: u1.OutputTokensDetails.ReasoningTokens + u2.OutputTokensDetails.ReasoningTokens,
		},
		TotalTokens: u1.TotalTokens + u2.TotalTokens,
	}
}
