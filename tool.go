package agentgo

import (
	"context"
	"log"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
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

func toolLoop(ctx context.Context, iteration int, params Params) (*models.ToolExecuteOutput, bool, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	messages := ctx.Value(models.MessagesContextKey).(*[]models.Message)
	runner := ctx.Value(models.RunnerContextKey).(*utils.Runner[*models.ToolExecuteOutput, models.ExecutionContext])
	execCtx := ctx.Value(models.ExecutionContextKey).(*models.ExecutionContext)

	log.Printf("Starting iteration %d", iteration)

	toolCalls, err := provider.ResolveToolCall(ctx, providers.AgentProviderPromptMessageParams{Messages: *messages}, params.Tools)

	switch {
	case err != nil:
		return nil, true, err
	case len(toolCalls) == 0:
		log.Printf("No tool calls resolved, generating text output for iteration %d", iteration)
		output := resolveTextOutputAsToolExecuteOutput(
			provider.GenerateText(ctx, providers.AgentProviderPromptMessageParams{
				Messages: *messages,
			}),
		)
		return &output, false, nil
	default:
		log.Printf("Resolved %d tool calls, executing them for iteration %d", len(toolCalls), iteration)
		utils.Each(toolCalls, func(toolCall models.ToolCall) {
			log.Printf("Executing tool %s", toolCall.ToolName)
			runner.Execute(
				"tool",
				func(iteration int, execCtx *models.ExecutionContext) (*models.ToolExecuteOutput, bool, error) {
					result, err := executeToolCall(ctx, toolCall, params)
					return result, true, err
				},
				iteration,
				execCtx,
			)
		})

		return nil, true, nil
	}
}
