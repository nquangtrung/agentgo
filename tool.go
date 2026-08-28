package agentgo

import (
	"context"
	"log"

	"trontria.com/agentgo/models"
)

func executeToolCall(_ context.Context, toolCall models.ToolCall, params Params) (*models.ToolExecuteOutput, error) {
	tool, err := resolveToolFromToolCall(toolCall, params.Tools)
	if err != nil {
		return nil, err
	}
	log.Printf("Resolved tool [%s]", tool.Name())
	toolResult := tool.Execute(models.ToolExecuteParams{
		Input: toolCall.Params,
	})
	log.Printf("Executed tool %s with result: %v", tool.Name(), toolResult)
	toolResult.ToolCall = &toolCall
	return &toolResult, nil
}
