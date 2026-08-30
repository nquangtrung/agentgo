package providers

import (
	"context"

	"trontria.com/agentgo/models"
)

type AgentProviderPromptMessageParams struct {
	Messages []models.Message
}

//go:generate mockgen -destination=../mocks/mock_agent_provider.go -package=mocks trontria.com/agentgo/providers AgentProvider
type AgentProvider interface {
	Context() models.LanguageModelContext
	GenerateText(ctx context.Context, params AgentProviderPromptMessageParams) (models.LanguageModelOutput, error)
	StreamText(ctx context.Context, params AgentProviderPromptMessageParams, emitter models.PartEmitter) (models.LanguageModelOutput, error)
	ResolveToolCall(ctx context.Context, params AgentProviderPromptMessageParams, toolParams []models.BaseTool) ([]models.ToolCall, error)
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
