package agentgo

import (
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

func StreamText(params Params) models.LanguageModelStreamOutput {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)

	channel := make(chan models.Part)
	go func() {
		defer close(channel)

		// for context := models.ExecutionContextFromLanguageModelContext(params.Provider.GetContext()); canProceedToNextStep(context, params); context = context.NextStep("tool step") {
		// 	channel <- models.NewStepStartPart(params.Provider.GetContext(), "tool step")

		// 	toolCalls := ResolveToolCalls(params, params.Provider)
		// 	executionOutputs := []models.ToolExecuteOutput{}
		// 	for _, tool := range toolCalls {
		// 		channel <- models.NewToolStartPart(params.Provider.GetContext(), tool.ToolName)
		// 		toolOutput, err := ResolveToolCallExecution(params, params.Provider, tool)
		// 		if err != nil {
		// 			channel <- models.NewToolErrorPart(params.Provider.GetContext(), tool.ToolName, map[string]any{
		// 				"error": err.Error(),
		// 			})
		// 			log.Printf("Error executing tool %s: %v", tool.ToolName, err)
		// 			continue
		// 		}

		// 		executionOutputs = append(executionOutputs, toolOutput)
		// 		channel <- models.NewToolResultPart(params.Provider.GetContext(), tool.ToolName, toolOutput.Result, models.LanguageModelUsage{
		// 			// TODO: Implement token counting for tool execution output
		// 			OutputTokens:    0,
		// 			InputTokens:     0,
		// 			CachedTokens:    0,
		// 			ReasoningTokens: 0,
		// 		})
		// 	}

		// 	channel <- models.NewStepEndPart(params.Provider.GetContext(), "tool step", models.LanguageModelUsage{
		// 		// TODO: Implement token counting for streaming output
		// 		OutputTokens:    0,
		// 		InputTokens:     0,
		// 		CachedTokens:    0,
		// 		ReasoningTokens: 0,
		// 	})
		// }

		provider.StreamText(providers.AgentProviderPromptMessageParams{
			Messages: messages,
		}, channel)
	}()

	return models.NewLanguageModelStreamOutput(channel, params.Provider.Context().ModelName)
}
