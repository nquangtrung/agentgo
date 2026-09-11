package fsm

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

type StartState struct {
}

func (s *StartState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)

	emitter.Emit(models.NewProcessStartPart(provider.Context()))
	fsmCtx.TotalUsage = models.LanguageModelUsage{}

	return &StepStartState{}, nil
}
