package agentgo

import (
	"log"

	"trontria.com/agentgo/models"
)

func accumulateToolCallResult(acc *models.ExecutionContext, item *models.ToolExecuteOutput, messages *[]models.Message) {
	log.Printf("Accumulator called with item: %+v", item)
	switch {
	case item == nil:
		log.Printf("Accumulator received nil item, skipping")
		return
	case item.ToolCall != nil:
		log.Printf("Adding tool call result to execution context: %s", item.ToolCall.ToolName)
		acc.AddStepWithResult(item.ToolCall.ToolName, item)
		*messages = append(*messages, models.NewMessageFromToolResult(*item))
		return
	default:
		log.Printf("Adding text output to execution context: %s", item.Output)
		acc.AddStepWithResult("text", item)
	}
}
