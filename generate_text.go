package agentgo

import (
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

func doLoop(provider providers.AgentProvider, params Params) (models.LanguageModelOutput, error) {
	context := models.NewExecutionContextFromLanguageModelContext(
		models.NewLanguageModelContext(params.ModelName),
	)

	messages := resolveMessages(params)
	for {
		canContinue := canProceedToNextStep(&context, params)
		if !canContinue {
			log.Printf("End conditions met, breaking the loop.")
			break
		}

		log.Printf("Resolving tool call")
		toolCall := provider.ResolveToolCall(
			providers.AgentProviderPromptMessageParams{
				Messages: messages,
			},
			params.Tools,
		)

		if toolCall == nil {
			log.Printf("No tool call was resolved, breaking the loop.")
			// No tool call was resolved, so we will break the loop and return the final output.
			context.AddStep("text", nil)
			output := resolveTextOutputAsToolExecuteOutput(
				provider.GenerateText(providers.AgentProviderPromptMessageParams{
					Messages: messages,
				}))
			context.UpdateLastStepResult(&output)
			break
		}

		context.AddStep(fmt.Sprintf("tool call [%s]", toolCall.ToolName), toolCall)
		tool, err := resolveToolFromToolCall(*toolCall, params.Tools)
		if err != nil {
			context.UpdateLastStepError(err)
			continue
		}
		log.Printf("Resolved tool [%s]", tool.Name())
		toolResult := tool.Execute(models.ToolExecuteParams{
			Params: toolCall.Params,
		})
		log.Printf("Executed tool %s with result: %v", tool.Name(), toolResult)
		context.UpdateLastStepResult(
			&toolResult,
		)

		log.Printf("Adding tool result to messages: %v", toolResult)
		messages = append(messages, models.NewMessageFromToolResult(toolResult))
	}
	return resolveExecutionContextAsTextOutput(&context), nil
}

func GenerateText(params Params) (models.LanguageModelOutput, error) {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)

	if len(params.EndConditions) == 0 || len(params.Tools) == 0 {
		// If no end conditions are provided, for safety, we will default to a max steps end condition of 1.
		return provider.GenerateText(providers.AgentProviderPromptMessageParams{
			Messages: messages,
		})
	}

	return doLoop(provider, params)
}
