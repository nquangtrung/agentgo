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
	ctx = context.WithValue(ctx, models.ToolMiddlewaresContextKey, params.ToolMiddlewares)
	ctx = context.WithValue(ctx, models.StepMiddlewaresContextKey, params.StepMiddlewares)

	config := graph.InvocationConfig[agentState, agentStateDelta]{
		Checkpointer: resolveGraphCheckpointer(params.Checkpointer),
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		if params.Resume {
			g, ok := retrieveGraphForResume(params.Checkpointer, params.ThreadID)
			if !ok {
				g = createGenerateTextGraph()
			}
			g.Resume(ctx, params.ThreadID, params.InterruptResults, config)
			return
		}

		g := createGenerateTextGraph()
		_, err := g.Invoke(ctx, agentState{
			toolExecutionsArchive: execArchive,
			messages:              messages,
		}, config)

		if err != nil {
			if superStepErr, ok := err.(*graph.SuperStepExecutionError); ok {
				interrupts := superStepErr.Interrupts()
				if len(interrupts) > 0 {
					storeGraphForResume(params.Checkpointer, interrupts[0].ThreadID, g)
				}
			}
		}
	})

	go func() {
		wg.Wait()
		close(partChannel)
	}()

	return models.NewLanguageModelStreamOutput(partChannel, params.Provider.Context().ModelName)
}
