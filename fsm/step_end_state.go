package fsm

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

type StepEndState struct {
	toEnd bool
}

func (s *StepEndState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
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
	if !s.toEnd {
		return &StepStartState{}, nil
	} else {
		return &EndState{}, nil
	}
}
