package agentgo

import (
	"context"
	"sync"

	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/models"
)

func StreamText(ctx context.Context, params Params) models.LanguageModelStreamOutput {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)
	execArchive := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	partChannel := make(chan models.Part)
	emitter := models.NewPartEmitter(partChannel)

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, params.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, params.Tools)
	ctx = context.WithValue(ctx, models.StreamContextKey, true)
	ctx = context.WithValue(ctx, models.PartEmitterContextKey, emitter)
	ctx = context.WithValue(ctx, models.PrepareStepFnContextKey, params.PrepareStep)

	var wg sync.WaitGroup
	wg.Go(func() {
		g := createGenerateTextGraph()
		config := graph.InvocationConfig[agentState, agentStateDelta]{}
		g.Invoke(ctx, agentState{
			toolExecutionsArchive: execArchive,
			messages:              messages,
		}, config)
	})

	go func() {
		wg.Wait()
		close(partChannel)
	}()

	return models.NewLanguageModelStreamOutput(partChannel, params.Provider.Context().ModelName)
}
