package agentgo

import (
	"context"

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
	messages := ctx.Value(models.MessagesContextKey).(*[]models.Message)

	shouldProceed := func(iteration int, execContext *models.ExecutionContext) bool {
		return canProceedToNextStep(execContext, params)
	}
	accumulator := func(acc *models.ExecutionContext, item *models.ToolExecuteOutput) {
		accumulateToolCallResult(acc, item, messages)
	}
	runner := utils.Runner[*models.ToolExecuteOutput, models.ExecutionContext]{
		Accumulator: accumulator,
	}

	ctx = context.WithValue(ctx, models.RunnerContextKey, &runner)
	loop := func(iteration int, execCtx *models.ExecutionContext) (*models.ToolExecuteOutput, bool, error) {
		return toolLoop(ctx, iteration, params)
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
	ctx = context.WithValue(ctx, models.MessagesContextKey, &messages)

	if len(params.EndConditions) == 0 || len(params.Tools) == 0 {
		// If no end conditions are provided, for safety, we will default to a max steps end condition of 1.
		return provider.GenerateText(ctx, providers.AgentProviderPromptMessageParams{
			Messages: messages,
		})
	}

	return doLoop(ctx, params)
}
