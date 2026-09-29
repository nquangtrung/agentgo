package agentgo

import (
	"context"
	"sync"

	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/models"
)

// StreamObject is the streaming counterpart of GenerateObject. It returns the
// same part channel as StreamText; the validated object arrives as a final
// models.ObjectPart (or models.ObjectErrorPart if every attempt failed).
func StreamObject[T any](ctx context.Context, params ObjectParams[T]) models.LanguageModelStreamOutput {
	baseParams := params.baseParams()
	provider := mustResolveProviderFromParams(baseParams)
	messages := resolveMessages(baseParams)
	execArchive := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	partChannel := make(chan models.Part)
	emitter := models.NewPartEmitter(partChannel)

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, baseParams.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, baseParams.Tools)
	ctx = context.WithValue(ctx, models.StreamContextKey, true)
	ctx = context.WithValue(ctx, models.PartEmitterContextKey, emitter)
	ctx = context.WithValue(ctx, models.PrepareStepFnContextKey, baseParams.PrepareStep)
	ctx = context.WithValue(ctx, models.ToolMiddlewaresContextKey, baseParams.ToolMiddlewares)
	ctx = context.WithValue(ctx, models.StepMiddlewaresContextKey, baseParams.StepMiddlewares)
	ctx = context.WithValue(ctx, models.SchemaContextKey, params.Schema)
	ctx = context.WithValue(ctx, models.MaxObjectRetriesContextKey, params.MaxObjectRetries)

	config := graph.InvocationConfig[agentState, agentStateDelta]{
		Checkpointer: resolveGraphCheckpointer(baseParams.Checkpointer),
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		if baseParams.Resume {
			g, ok := retrieveGraphForResume(baseParams.Checkpointer, baseParams.ThreadID)
			if !ok {
				g = createAgentGraph()
			}
			g.Resume(ctx, baseParams.ThreadID, baseParams.InterruptResults, config)
			return
		}

		g := createAgentGraph()
		_, err := g.Invoke(ctx, agentState{
			toolExecutionsArchive: execArchive,
			messages:              messages,
		}, config)

		if err != nil {
			if superStepErr, ok := err.(*graph.SuperStepExecutionError); ok {
				interrupts := superStepErr.Interrupts()
				if len(interrupts) > 0 {
					storeGraphForResume(baseParams.Checkpointer, interrupts[0].ThreadID, g)
				}
			}
		}
	})

	go func() {
		wg.Wait()
		close(partChannel)
	}()

	return models.NewLanguageModelStreamOutput(partChannel, provider.Context().ModelName)
}
