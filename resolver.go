package agentgo

import (
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

func NewMessageFromTools(toolOutputs []models.ToolExecuteOutput) []models.Message {
	messages := []models.Message{}
	for _, toolOutput := range toolOutputs {
		messages = append(messages, models.NewMessageFromToolResult(toolOutput))
	}
	return messages
}

func ResolveMessages(params Params) []models.Message {
	messages := params.Messages
	if params.Prompt != "" {
		messages = append(messages, models.NewHumanStringMessage(params.Prompt))
	}

	return messages
}

func ResolveToolCalls(params Params, provider providers.AgentProvider) []models.ToolCall {
	messages := ResolveMessages(params)
	toolCalls, err := provider.ResolveToolCalls(providers.AgentProviderPromptMessageParams{
		Messages: messages,
	}, params.Tools)

	if err != nil {
		// Handle the error appropriately, e.g., log it or send it through the channel
		return []models.ToolCall{}
	}
	return toolCalls
}

func ResolveToolCallExecution(params Params, provider providers.AgentProvider, tool models.ToolCall) (models.ToolExecuteOutput, error) {
	return models.ToolExecuteOutput{}, nil
}
