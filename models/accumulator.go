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
		InputTokens:     u1.InputTokens + u2.InputTokens,
		OutputTokens:    u1.OutputTokens + u2.OutputTokens,
		CachedTokens:    u1.CachedTokens + u2.CachedTokens,
		ReasoningTokens: u1.ReasoningTokens + u2.ReasoningTokens,
	}
}
