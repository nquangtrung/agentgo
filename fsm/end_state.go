package fsm

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

type EndState struct {
}

func (s *EndState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)

	emitter.Emit(models.NewProcessEndPart(provider.Context(), fsmCtx.TotalUsage, models.FinishReasonCompleted))

	return nil, nil
}
