package providers

import (
	"trontria.com/agentgo/models"
)

type AgentProviderPromptMessageParams struct {
	Messages []models.Message
}

//go:generate mockgen -destination=../mocks/mock_agent_provider.go -package=mocks trontria.com/agentgo/providers AgentProvider
type AgentProvider interface {
	Context() models.LanguageModelContext
	GenerateText(params AgentProviderPromptMessageParams) (models.LanguageModelOutput, error)
	StreamText(params AgentProviderPromptMessageParams, channel chan models.Part)
	ResolveToolCall(params AgentProviderPromptMessageParams, toolParams []models.Tool) *models.ToolCall
}

type BaseAgentProvider struct {
	context models.LanguageModelContext
}

func (p BaseAgentProvider) Context() models.LanguageModelContext {
	return p.context
}

func NewBaseAgentProvider(context models.LanguageModelContext) BaseAgentProvider {
	return BaseAgentProvider{
		context: context,
	}
}
