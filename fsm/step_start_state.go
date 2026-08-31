package fsm

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

type StepStartState struct {
}

func (s *StepStartState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	prepareStep := ctx.Value(models.PrepareStepFnContextKey).(PrepareStepFn)

	emitter.Emit(models.NewStepStartPart(
		provider.Context(),
		"step",
	))

	step := Step{
		StepIndex: len(fsmCtx.Steps) + 1,
	}

	if prepareStep != nil {
		prepareStepResult, _ := prepareStep(step, *fsmCtx)
		step.PrepareStepResult = prepareStepResult
	}

	fsmCtx.CurrentStep = step
	fsmCtx.Steps = append(fsmCtx.Steps, step)

	return &PredicateState{}, nil
}
