package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

func prepareProcess(ctx context.Context, state agentState) (agentStateDelta, error) {
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter.Emit(models.NewProcessStartPart(provider.Context()))

	return agentStateDelta{
		from: PREPARE_PROCESS,
	}, nil
}

func prepareStep(ctx context.Context, state agentState) (agentStateDelta, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	prepareStep := ctx.Value(models.PrepareStepFnContextKey).(PrepareStepFn)

	emitter.Emit(models.NewStepStartPart(
		provider.Context(),
		"step",
	))

	step := step{
		index: len(state.steps) + 1,
	}
	if prepareStep != nil {
		prepareStepResult, _ := prepareStep(step, state)
		step.prepareStepResult = prepareStepResult
	}

	// Run StepMiddleware Before hooks
	stepMiddlewares := ctx.Value(models.StepMiddlewaresContextKey).([]StepMiddleware)
	for _, mw := range stepMiddlewares {
		if mw.Before != nil {
			mwCtx := buildStepMiddlewareContext(state, step)
			err := mw.Before(ctx, mwCtx)
			if err != nil {
				logger.Warn("Step middleware Before hook error", "stepIndex", step.index, "error", err)
				// Any error from step middleware aborts the pipeline
				return agentStateDelta{}, err
			}
		}
	}

	return agentStateDelta{
		from:    PREPARE_STEP,
		addStep: &step,
	}, nil
}

func endProcess(ctx context.Context, state agentState) (agentStateDelta, error) {
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)

	emitter.Emit(models.NewProcessEndPart(provider.Context(), state.totalUsage, models.FinishReasonCompleted))
	return agentStateDelta{}, nil
}

func endStep(ctx context.Context, state agentState) (agentStateDelta, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)

	// Run StepMiddleware After hooks before emitting StepEndPart
	stepMiddlewares := ctx.Value(models.StepMiddlewaresContextKey).([]StepMiddleware)
	for _, mw := range stepMiddlewares {
		if mw.After != nil {
			mwCtx := buildStepMiddlewareContext(state, state.currentStep)
			err := mw.After(ctx, mwCtx)
			if err != nil {
				logger.Warn("Step middleware After hook error", "stepIndex", state.currentStep.index, "error", err)
				// Any error from step middleware aborts the pipeline
				return agentStateDelta{}, err
			}
		}
	}

	// TODO handle usage accumulate
	emitter.Emit(models.NewStepEndPart(
		provider.Context(),
		"step",
		state.currentStep.usage,
	))
	return agentStateDelta{}, nil
}
