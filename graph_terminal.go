package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

func prepareProcess(ctx context.Context, state agentState) (agentState, error) {
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter.Emit(models.NewProcessStartPart(provider.Context()))

	return agentState{
		toolExecutionsArchive: &models.ToolExecutionsArchive{},
		messages:              &[]models.Message{},
		totalUsage:            models.LanguageModelUsage{},
	}, nil
}

func prepareStep(ctx context.Context, state agentState) (agentState, error) {
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

	return agentState{
		currentStep: step,
	}, nil
}

func endProcess(ctx context.Context, state agentState) (agentState, error) {
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)

	emitter.Emit(models.NewProcessEndPart(provider.Context(), state.totalUsage, models.FinishReasonCompleted))
	return state, nil
}

func endStep(ctx context.Context, state agentState) (agentState, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	// TODO handle usage accumulate
	emitter.Emit(models.NewStepEndPart(
		provider.Context(),
		"step",
		state.currentStep.usage,
	))
	return state, nil
}
