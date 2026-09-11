package providers

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
)

type AgentProviderPromptMessageParams struct {
	Messages []models.Message
}

//go:generate mockgen -destination=../mocks/mock_agent_provider.go -package=mocks github.com/nquangtrung/agentgo/providers AgentProvider
type AgentProvider interface {
	Context() models.LanguageModelContext
	GenerateText(ctx context.Context, params AgentProviderPromptMessageParams) (models.LanguageModelOutput, error)
	StreamText(ctx context.Context, params AgentProviderPromptMessageParams, emitter models.PartEmitter) (models.LanguageModelOutput, error)
	ResolveToolCall(ctx context.Context, params AgentProviderPromptMessageParams, toolParams []models.BaseTool) (models.LanguageModelToolCallResolveOutput, error)
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
