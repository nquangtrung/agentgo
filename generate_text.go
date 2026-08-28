package agentgo

import (
	"context"
	"log"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

func canProceedToNextStep(context *models.ExecutionContext, params Params) bool {
	conditions := params.EndConditions
	for _, condition := range conditions {
		if condition.Condition(context) {
			return false
		}
	}
	return true
}

func doLoop(ctx context.Context, params Params) (models.LanguageModelOutput, error) {
	execContext := ctx.Value(models.ExecutionContextKey).(*models.ExecutionContext)
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	messages := resolveMessages(params)
	ended := false

	shouldProceed := func(iteration int, execContext *models.ExecutionContext) bool {
		return canProceedToNextStep(execContext, params) && !ended
	}
	accumulator := func(acc *models.ExecutionContext, item *models.ToolExecuteOutput) {
		log.Printf("Accumulator called with item: %+v", item)
		switch {
		case item == nil:
			log.Printf("Accumulator received nil item, skipping")
			return
		case item.ToolCall != nil:
			log.Printf("Adding tool call result to execution context: %s", item.ToolCall.ToolName)
			acc.AddStepWithResult(item.ToolCall.ToolName, item)
			messages = append(messages, models.NewMessageFromToolResult(*item))
			return
		default:
			log.Printf("Adding text output to execution context: %s", item.Output)
			acc.AddStepWithResult("text", item)
		}
	}
	runner := utils.Runner[*models.ToolExecuteOutput, models.ExecutionContext]{
		Accumulator: accumulator,
	}

	loop := func(iteration int, execCtx *models.ExecutionContext) (*models.ToolExecuteOutput, error) {
		log.Printf("Starting iteration %d", iteration)

		toolCalls, err := provider.ResolveToolCall(ctx, providers.AgentProviderPromptMessageParams{Messages: messages}, params.Tools)

		switch {
		case err != nil:
			return nil, err
		case len(toolCalls) == 0:
			log.Printf("No tool calls resolved, generating text output for iteration %d", iteration)
			output := resolveTextOutputAsToolExecuteOutput(
				provider.GenerateText(ctx, providers.AgentProviderPromptMessageParams{
					Messages: messages,
				}),
			)
			ended = true
			return &output, nil
		default:
			log.Printf("Resolved %d tool calls, executing them for iteration %d", len(toolCalls), iteration)
			utils.Each(toolCalls, func(toolCall models.ToolCall) {
				log.Printf("Executing tool %s", toolCall.ToolName)
				runner.Execute(
					"tool",
					func(iteration int, execCtx *models.ExecutionContext) (*models.ToolExecuteOutput, error) {
						return executeToolCall(ctx, toolCall, params)
					},
					iteration,
					execCtx,
				)
			})

			return nil, nil
		}
	}

	runner.Loop("step", shouldProceed, loop, execContext, false)

	return resolveExecutionContextAsTextOutput(execContext)
}

func GenerateText(ctx context.Context, params Params) (models.LanguageModelOutput, error) {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)
	execContext := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	ctx = context.WithValue(ctx, models.ExecutionContextKey, execContext)
	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)

	if len(params.EndConditions) == 0 || len(params.Tools) == 0 {
		// If no end conditions are provided, for safety, we will default to a max steps end condition of 1.
		return provider.GenerateText(ctx, providers.AgentProviderPromptMessageParams{
			Messages: messages,
		})
	}

	return doLoop(ctx, params)
}
