package agentgo

import (
	"log"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

func GenerateText(params Params) (models.LanguageModelOutput, error) {
	if params.ModelName != "" {
		var provider, err = providers.CreateAgentProvider(providers.AgentProviderFactoryParams{
			ModelName: params.ModelName,
		})
		if err != nil {
			return models.LanguageModelOutput{}, err
		}
		params.Provider = provider
	}

	if params.Provider == nil {
		return models.LanguageModelOutput{}, &models.UnsupportedModelError{ModelName: "nil provider"}
	}

	messages := ResolveMessages(params)
	for context := models.ExecutionContextFromLanguageModelContext(params.Provider.GetContext()); canProceedToNextStep(context, params); context = context.NextStep("tool step") {
		toolCalls := ResolveToolCalls(params, params.Provider)

		executionOutputs := []models.ToolExecuteOutput{}
		for _, tool := range toolCalls {
			toolOutputs, err := ResolveToolCallExecution(params, params.Provider, tool)
			if err != nil {
				// XXX: Handle the error appropriately, e.g., log it or send it through the channel
				continue
			}

			executionOutputs = append(executionOutputs, toolOutputs)

		}
		messages = append(messages, NewMessageFromTools(executionOutputs)...)
		context = context.NextStep("tool step")
	}

	return params.Provider.GenerateText(providers.AgentProviderGenerateTextParams{
		AgentProviderPromptMessageParams: providers.AgentProviderPromptMessageParams{
			Messages: messages,
		},
	})
}

func canProceedToNextStep(context models.ExecutionContext, params Params) bool {
	return context.GetStepIndex() < params.MaxToolSteps
}

func StreamText(params Params) models.LanguageModelStreamOutput {
	if params.ModelName != "" {
		var provider, err = providers.CreateAgentProvider(providers.AgentProviderFactoryParams{
			ModelName: params.ModelName,
		})
		if err != nil {
			panic(err)
		}
		params.Provider = provider
	}

	if params.Provider == nil {
		panic("nil provider")
	}

	messages := ResolveMessages(params)

	channel := make(chan models.Part)
	go func() {
		defer close(channel)

		for context := models.ExecutionContextFromLanguageModelContext(params.Provider.GetContext()); canProceedToNextStep(context, params); context = context.NextStep("tool step") {
			channel <- models.NewStepStartPart(params.Provider.GetContext(), "tool step")

			toolCalls := ResolveToolCalls(params, params.Provider)
			executionOutputs := []models.ToolExecuteOutput{}
			for _, tool := range toolCalls {
				channel <- models.NewToolStartPart(params.Provider.GetContext(), tool.ToolName)
				toolOutput, err := ResolveToolCallExecution(params, params.Provider, tool)
				if err != nil {
					channel <- models.NewToolErrorPart(params.Provider.GetContext(), tool.ToolName, map[string]interface{}{
						"error": err.Error(),
					})
					log.Printf("Error executing tool %s: %v", tool.ToolName, err)
					continue
				}

				executionOutputs = append(executionOutputs, toolOutput)
				channel <- models.NewToolResultPart(params.Provider.GetContext(), tool.ToolName, toolOutput.Result, models.LanguageModelUsage{
					// TODO: Implement token counting for tool execution output
					OutputTokens:    0,
					InputTokens:     0,
					CachedTokens:    0,
					ReasoningTokens: 0,
				})
			}

			channel <- models.NewStepEndPart(params.Provider.GetContext(), "tool step", models.LanguageModelUsage{
				// TODO: Implement token counting for streaming output
				OutputTokens:    0,
				InputTokens:     0,
				CachedTokens:    0,
				ReasoningTokens: 0,
			})
		}

		params.Provider.StreamText(providers.AgentProviderStreamTextParams{
			AgentProviderPromptMessageParams: providers.AgentProviderPromptMessageParams{
				Messages: messages,
			},
		}, channel)
	}()

	return models.NewLanguageModelStreamOutput(channel, params.Provider.GetContext().ModelName)
}
