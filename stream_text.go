package agentgo

import (
	"context"
	"log"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

func doStreamLoop(ctx context.Context, params Params, channel chan models.Part) {
	execContext := ctx.Value(models.ExecutionContextKey).(*models.ExecutionContext)
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	messages := resolveMessages(params)

	log.Printf("Starting loop with context %v", execContext.ModelName())
	for {
		canContinue := canProceedToNextStep(execContext, params)
		if !canContinue {
			log.Printf("End conditions met, breaking the loop.")
			break
		}

		wrapStep(ctx, "step", channel, func() error {
			log.Printf("Resolving tool call")
			toolCalls, err := provider.ResolveToolCall(ctx, providers.AgentProviderPromptMessageParams{Messages: messages}, params.Tools)

			switch {
			case err != nil:
				log.Printf("Error resolving tool call: %v", err)
				return err
			case len(toolCalls) == 0:
				log.Printf("No tool call was resolved, breaking the loop.")

				// No tool call was resolved, so we will break the loop and return the final output.
				provider.StreamText(ctx, providers.AgentProviderPromptMessageParams{Messages: messages}, channel)
				return nil
			default:
				for _, toolCall := range toolCalls {
					wrapToolCall(ctx, toolCall.ToolName, channel, func() (*models.ToolExecuteOutput, error) {
						return executeToolCall(ctx, toolCall, params)
					})
				}
			}

			return nil
		})
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
