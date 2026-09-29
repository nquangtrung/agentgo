package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/models"
)

// GenerateObject asks the model for a structured object matching params.Schema,
// validates it, and retries with a repair prompt on failure.
func GenerateObject[T any](ctx context.Context, params ObjectParams[T]) (ObjectResult[T], error) {
	baseParams := params.baseParams()
	provider := mustResolveProviderFromParams(baseParams)
	emitter := models.NewEmptyPartEmitter()

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, baseParams.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, baseParams.Tools)
	ctx = context.WithValue(ctx, models.StreamContextKey, false)
	ctx = context.WithValue(ctx, models.PartEmitterContextKey, emitter)
	ctx = context.WithValue(ctx, models.PrepareStepFnContextKey, baseParams.PrepareStep)
	ctx = context.WithValue(ctx, models.ToolMiddlewaresContextKey, baseParams.ToolMiddlewares)
	ctx = context.WithValue(ctx, models.StepMiddlewaresContextKey, baseParams.StepMiddlewares)
	ctx = context.WithValue(ctx, models.SchemaContextKey, params.Schema)
	ctx = context.WithValue(ctx, models.MaxObjectRetriesContextKey, params.MaxObjectRetries)

	config := graph.InvocationConfig[agentState, agentStateDelta]{
		Checkpointer: resolveGraphCheckpointer(baseParams.Checkpointer),
	}

	if baseParams.Resume {
		g, ok := retrieveGraphForResume(baseParams.Checkpointer, baseParams.ThreadID)
		if !ok {
			g = createAgentGraph()
		}
		result, err := g.Resume(ctx, baseParams.ThreadID, baseParams.InterruptResults, config)
		if err != nil {
			return ObjectResult[T]{}, err
		}
		return resolveObjectResult[T](result)
	}

	messages := resolveMessages(baseParams)
	execArchive := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	g := createAgentGraph()
	result, err := g.Invoke(ctx, agentState{
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
		return ObjectResult[T]{}, err
	}

	return resolveObjectResult[T](result)
}

// resolveObjectResult pulls the validated object out of the execution context
// and asserts it to T.
func resolveObjectResult[T any](result agentState) (ObjectResult[T], error) {
	// When every attempt failed, nothing was archived — surface the accumulated
	// error rather than an empty-context error.
	if result.objectError != nil {
		return ObjectResult[T]{}, &models.NoObjectGeneratedError{
			Message:   result.objectError.Error(),
			Text:      result.objectRaw,
			Usage:     result.totalUsage,
			ModelName: result.toolExecutionsArchive.ModelName(),
			Cause:     result.objectError,
		}
	}

	object, raw, err := resolveExecutionContextAsObjectResult(result.toolExecutionsArchive)
	if err != nil {
		return ObjectResult[T]{}, err
	}

	typed, ok := object.(T)
	if !ok {
		return ObjectResult[T]{}, &models.ExecutionContextError{
			Message: "object does not match the schema's type",
		}
	}

	return ObjectResult[T]{
		Object:    typed,
		Raw:       raw,
		Usage:     result.totalUsage,
		ModelName: result.toolExecutionsArchive.ModelName(),
		Context:   result.toolExecutionsArchive,
	}, nil
}
