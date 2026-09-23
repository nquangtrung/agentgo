package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

func prepareText(ctx context.Context, state agentState) (agentState, error) {
	stream := ctx.Value(models.StreamContextKey).(bool)

	currentStep := state.currentStep
	currentStep.stream = stream
	return agentState{
		currentStep: currentStep,
	}, nil
}

func endText(_ context.Context, state agentState) (agentState, error) {
	// TODO handle usage accumulate
	return agentState{
		currentStep:   state.currentStep,
		textGenerated: true,
	}, nil
}

func generateText(ctx context.Context, state agentState) (agentState, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	messages := state.messages

	output, err := provider.GenerateText(ctx, providers.AgentProviderPromptMessageParams{
		Messages: *messages,
	})

	if err != nil {
		// Handle error and retry
		return state, err
	}

	currentStep := state.currentStep
	currentStep.usage = models.AccumulateUsage(currentStep.usage, output.Usage)
	state.totalUsage = models.AccumulateUsage(state.totalUsage, output.Usage)

	outputAsToolExecuteOutput := resolveTextOutputAsToolExecuteOutput(output, err)
	models.AccumulateToolCallResult(state.toolExecutionsArchive, &outputAsToolExecuteOutput, messages)
	return agentState{
		currentStep:           currentStep,
		toolExecutionsArchive: state.toolExecutionsArchive,
	}, nil
}

func streamText(ctx context.Context, state agentState) (agentState, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	messages := state.messages

	output, err := provider.StreamText(
		ctx,
		providers.AgentProviderPromptMessageParams{Messages: *messages},
		*emitter,
	)

	if err != nil {
		// TODO Handle error and retry
		return state, err
	}

	currentStep := state.currentStep
	currentStep.usage = models.AccumulateUsage(currentStep.usage, output.Usage)
	state.totalUsage = models.AccumulateUsage(state.totalUsage, output.Usage)

	outputAsToolExecuteOutput := resolveTextOutputAsToolExecuteOutput(output, err)
	models.AccumulateToolCallResult(state.toolExecutionsArchive, &outputAsToolExecuteOutput, messages)
	return agentState{
		currentStep:           currentStep,
		toolExecutionsArchive: state.toolExecutionsArchive,
	}, nil
}
