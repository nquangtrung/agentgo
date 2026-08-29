package agentgo

import (
	"context"
	"sync"

	"trontria.com/agentgo/fsm"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"
)

func handleEvent(ctx context.Context, event utils.RunnerEvent[*models.ToolExecuteOutput, models.ExecutionContext], partChannel chan models.Part, provider models.LanguageModelContext) {
	switch event.Type {
	case utils.RunnerEventStart:
		switch event.ActionName {
		case "tool":
			partChannel <- models.NewToolStartPart(provider, event.ActionName)
		case "step":
			partChannel <- models.NewStepStartPart(provider, event.ActionName)
		}
		// case utils.RunnerEventEnd:
		// 	partChannel <- models.NewToolEndPart(provider.Context(), event.ActionName)
	}
}

func StreamText(ctx context.Context, params Params) models.LanguageModelStreamOutput {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)
	execContext := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	partChannel := make(chan models.Part)
	accumulator := func(acc *models.ExecutionContext, item *models.ToolExecuteOutput) {
		accumulateToolCallResult(acc, item, &messages)
	}

	runner := utils.NewRunner(accumulator)
	machine := fsm.New[fsm.AgentContext]()

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.RunnerContextKey, runner)
	ctx = context.WithValue(ctx, models.MachineContextKey, machine)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, params.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, params.Tools)
	ctx = context.WithValue(ctx, models.StreamContextKey, true)
	ctx = context.WithValue(ctx, models.StreamPartChannelContextKey, partChannel)

	var wg sync.WaitGroup
	wg.Go(func() {
		for event := range runner.Channel() {
			handleEvent(ctx, event, partChannel, provider.Context())
			// switch event.Type {
			// case utils.RunnerEventStart:
			// 	switch event.ActionName {
			// 	case "tool":
			// 		partChannel <- models.NewToolStartPart(provider.Context(), event.ActionName)
			// 	case "step":
			// 		partChannel <- models.NewStepStartPart(provider.Context(), event.ActionName)
			// 	}
			// 	// case utils.RunnerEventEnd:
			// 	// 	partChannel <- models.NewToolEndPart(provider.Context(), event.ActionName)
			// }
		}
	})
	wg.Go(func() {
		agentCtx := &fsm.AgentContext{
			Messages:         &messages,
			ExecutionContext: execContext,
		}
		machine.Run(ctx, &fsm.StartState{}, agentCtx)
	})
	go func() {
		wg.Wait()
		close(partChannel)
	}()

	return models.NewLanguageModelStreamOutput(partChannel, params.Provider.Context().ModelName)
}
