package agentgo

import (
	"context"
	"fmt"
	"log"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
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

	log.Printf("Starting loop with context %v", execContext.ModelName())
	messages := resolveMessages(params)
	for {
		canContinue := canProceedToNextStep(execContext, params)
		if !canContinue {
			log.Printf("End conditions met, breaking the loop.")
			break
		}

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
			execContext.AddStep("text", nil)
			output := resolveTextOutputAsToolExecuteOutput(
				provider.GenerateText(ctx, providers.AgentProviderPromptMessageParams{
					Messages: messages,
				}))
			execContext.UpdateLastStepResult(&output)
			break
		}

		for _, toolCall := range toolCalls {
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

			log.Printf("Adding tool result to messages: %v", toolResult)
			messages = append(messages, models.NewMessageFromToolResult(toolResult))
		}
	}

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
