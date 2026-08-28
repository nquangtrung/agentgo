package agentgo

import (
	"context"
	"log"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

func wrapStep(ctx context.Context, stepName string, channel chan models.Part, fn func() error) error {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	stepStartPart := models.NewStepStartPart(provider.Context(), stepName)
	channel <- stepStartPart

	err := fn()
	if err != nil {
		log.Printf("Error in step %s: %v", stepName, err)

		stepErrorPart := models.NewStepErrorPart(
			provider.Context(),
			stepName,
			models.NewLanguageModelUsage(0, 0, 0, 0),
			err,
		)
		channel <- stepErrorPart

		return err
	}

	execContext := ctx.Value(models.ExecutionContextKey).(*models.ExecutionContext)
	stepEndPart := models.NewStepEndPart(provider.Context(), stepName, execContext.LastStep().ToolResult.Usage)
	channel <- stepEndPart
	return nil
}

func wrapToolCall(ctx context.Context, toolName string, channel chan models.Part, fn func() (*models.ToolExecuteOutput, error)) error {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	toolStartPart := models.NewToolStartPart(provider.Context(), toolName)
	channel <- toolStartPart

	toolResult, err := fn()
	if err != nil {
		log.Printf("Error in tool call %s: %v", toolName, err)
		toolErrorPart := models.NewToolErrorPart(provider.Context(), toolName, models.NewLanguageModelUsage(0, 0, 0, 0), err)
		channel <- toolErrorPart
		return err
	}

	toolResultPart := models.NewToolResultPart(provider.Context(), toolName, *toolResult)
	channel <- toolResultPart

	return nil
}

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
	return &toolResult, nil
}

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
