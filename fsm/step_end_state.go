package fsm

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

type StepEndState struct {
}

func finalizeStep(ctx context.Context, fsmCtx *AgentContext) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	emitter.Emit(models.NewStepEndPart(
		provider.Context(),
		"step",
		fsmCtx.CurrentStep.Usage,
	))

	fsmCtx.TotalUsage = models.AccumulateUsage(
		fsmCtx.TotalUsage,
		fsmCtx.CurrentStep.Usage,
	)
}

func (s *StepEndState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	finalizeStep(ctx, fsmCtx)
	return &StepStartState{}, nil
}
