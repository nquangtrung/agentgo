package agentgo

import (
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

func newMessageFromTools(toolOutputs []models.ToolExecuteOutput) []models.Message {
	messages := []models.Message{}
	for _, toolOutput := range toolOutputs {
		messages = append(messages, models.NewMessageFromToolResult(toolOutput))
	}
	return messages
}

func resolveMessages(params Params) []models.Message {
	messages := params.Messages
	if params.Prompt != "" {
		messages = append(messages, models.NewHumanStringMessage(params.Prompt))
	}

	return messages
}

func resolveToolCallExecution(params Params, provider providers.AgentProvider, tool models.ToolCall) (models.ToolExecuteOutput, error) {
	return models.ToolExecuteOutput{}, nil
}

func resolveProviderFromParams(params Params) (providers.AgentProvider, error) {
	if params.ModelName != "" {
		var provider, err = CreateAgentProvider(AgentProviderFactoryParams{
			ModelName: params.ModelName,
		})
		if err != nil {
			return nil, err
		}
		params.Provider = provider
	}

	if params.Provider == nil {
		return nil, &models.UnsupportedModelError{ModelName: "nil provider"}
	}

	return params.Provider, nil
}

func mustResolveProviderFromParams(params Params) providers.AgentProvider {
	return utils.Must(resolveProviderFromParams(params))
}

func resolveToolFromToolCall(toolCall models.ToolCall, tools []models.Tool) (models.Tool, error) {
	for _, tool := range tools {
		if tool.Name() == toolCall.ToolName {
			return tool, nil
		}
	}
	return nil, &models.ToolNotFoundError{ToolName: toolCall.ToolName}
}

func resolveTextOutputAsToolExecuteOutput(textOutput models.LanguageModelOutput, err error) models.ToolExecuteOutput {
	if err != nil {
		return models.ToolExecuteOutput{
			Result: nil,
			Error:  err,
			Usage:  textOutput.Usage,
		}
	}
	return models.ToolExecuteOutput{
		Result: map[string]any{
			"text": textOutput.Text,
		},
		Error: nil,
		Usage: textOutput.Usage,
	}
}

func resolveExecutionContextAsTextOutput(context *models.ExecutionContext) models.LanguageModelOutput {
	return models.LanguageModelOutput{
		Text:    context.LastStep().ToolResult.Result["text"].(string),
		Usage:   context.LastStep().ToolResult.Usage,
		Context: context,
	}
}
