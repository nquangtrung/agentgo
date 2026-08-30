package agentgo

import (
	"context"
	"sync"

	"trontria.com/agentgo/fsm"
	"trontria.com/agentgo/models"
)

func StreamText(ctx context.Context, params Params) models.LanguageModelStreamOutput {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)
	execContext := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	partChannel := make(chan models.Part)

	machine := fsm.New[fsm.AgentContext]()
	emitter := models.NewPartEmitter(partChannel)

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.MachineContextKey, machine)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, params.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, params.Tools)
	ctx = context.WithValue(ctx, models.StreamContextKey, true)
	ctx = context.WithValue(ctx, models.PartEmitterContextKey, emitter)

	var wg sync.WaitGroup
	wg.Go(func() {
		agentCtx := &fsm.AgentContext{
			Messages:              &messages,
			ToolExecutionsArchive: execContext,
		}
		machine.Run(ctx, &fsm.StartState{}, agentCtx)
	})
	go func() {
		wg.Wait()
		close(partChannel)
	}()

	return models.NewLanguageModelStreamOutput(partChannel, params.Provider.Context().ModelName)
}
