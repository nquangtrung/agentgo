package agentgo

import (
	"context"

	"trontria.com/agentgo/fsm"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"
)

func StreamText(ctx context.Context, params Params) models.LanguageModelStreamOutput {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)
	execContext := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	partChannel := make(chan models.Part)
	eventChannel := make(chan utils.RunnerEvent[*models.ToolExecuteOutput, models.ExecutionContext])
	accumulator := func(acc *models.ExecutionContext, item *models.ToolExecuteOutput) {
		accumulateToolCallResult(acc, item, &messages)
	}

	runner := utils.NewRunner(eventChannel, accumulator)
	machine := fsm.New[fsm.AgentContext]()

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.RunnerContextKey, runner)
	ctx = context.WithValue(ctx, models.MachineContextKey, machine)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, params.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, params.Tools)
	ctx = context.WithValue(ctx, models.StreamContextKey, true)
	ctx = context.WithValue(ctx, models.StreamPartChannelContextKey, partChannel)

	go func() {
		defer close(partChannel)
		defer close(eventChannel)

		agentCtx := &fsm.AgentContext{
			Messages:         &messages,
			ExecutionContext: execContext,
		}

		machine.Run(ctx, &fsm.PredicateState{}, agentCtx)
	}()

	return models.NewLanguageModelStreamOutput(partChannel, params.Provider.Context().ModelName)
}
