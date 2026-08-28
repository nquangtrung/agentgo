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
