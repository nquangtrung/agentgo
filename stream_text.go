package agentgo

import (
	"context"
	"fmt"
	"log"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

func doStreamLoop(ctx context.Context, params Params, channel chan models.Part) {
	execContext := ctx.Value(models.ExecutionContextKey).(*models.ExecutionContext)
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)

	log.Printf("Starting loop with context %v", execContext.ModelName())
	messages := resolveMessages(params)
	for {
		canContinue := canProceedToNextStep(execContext, params)
		if !canContinue {
			log.Printf("End conditions met, breaking the loop.")
			break
		}

		stepStartPart := models.NewStepStartPart(provider.Context(), "step start")
		channel <- stepStartPart

		log.Printf("Resolving tool call")
		toolCalls, err := provider.ResolveToolCall(
			ctx,
			providers.AgentProviderPromptMessageParams{
				Messages: messages,
			},
			params.Tools,
		)

		if err != nil {
			log.Printf("Error resolving tool call: %v", err)
			execContext.UpdateLastStepError(err)
			continue
		}

		if len(toolCalls) == 0 {
			log.Printf("No tool call was resolved, breaking the loop.")

			// No tool call was resolved, so we will break the loop and return the final output.
			provider.StreamText(
				ctx,
				providers.AgentProviderPromptMessageParams{
					Messages: messages,
				},
				channel,
			)

			stepEndPart := models.NewStepEndPart(provider.Context(), "step end", execContext.LastStep().ToolResult.Usage)
			channel <- stepEndPart
			break
		}

		for _, toolCall := range toolCalls {
			toolStartPart := models.NewToolStartPart(provider.Context(), toolCall.ToolName)
			channel <- toolStartPart

			execContext.AddStep(fmt.Sprintf("tool call [%s]", toolCall.ToolName), &toolCall)
			tool, err := resolveToolFromToolCall(toolCall, params.Tools)
			if err != nil {
				execContext.UpdateLastStepError(err)
				continue
			}
			log.Printf("Resolved tool [%s]", tool.Name())
			toolResult := tool.Execute(models.ToolExecuteParams{
				Input: toolCall.Params,
			})
			log.Printf("Executed tool %s with result: %v", tool.Name(), toolResult)
			execContext.UpdateLastStepResult(
				&toolResult,
			)

			toolResultPart := models.NewToolResultPart(provider.Context(), toolCall.ToolName, toolResult)
			channel <- toolResultPart

			log.Printf("Adding tool result to messages: %v", toolResult)
			messages = append(messages, models.NewMessageFromToolResult(toolResult))
		}

		stepEndPart := models.NewStepEndPart(provider.Context(), "step end", execContext.LastStep().ToolResult.Usage)
		channel <- stepEndPart
	}
}

func StreamText(ctx context.Context, params Params) models.LanguageModelStreamOutput {
	provider := mustResolveProviderFromParams(params)

	execCtx := models.NewExecutionContextFromLanguageModelContext(provider.Context())
	ctx = context.WithValue(ctx, models.ExecutionContextKey, execCtx)
	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)

	channel := make(chan models.Part)
	go func() {
		defer close(channel)
		if len(params.EndConditions) == 0 || len(params.Tools) == 0 {
			messages := resolveMessages(params)
			provider.StreamText(ctx, providers.AgentProviderPromptMessageParams{Messages: messages}, channel)
			return
		}

		doStreamLoop(ctx, params, channel)
	}()

	return models.NewLanguageModelStreamOutput(channel, params.Provider.Context().ModelName)
}
