package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/models"
)

func GenerateText(ctx context.Context, params Params) (models.LanguageModelOutput, error) {
	provider := mustResolveProviderFromParams(params)
	emitter := models.NewEmptyPartEmitter()

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, params.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, params.Tools)
	ctx = context.WithValue(ctx, models.StreamContextKey, false)
	ctx = context.WithValue(ctx, models.PartEmitterContextKey, emitter)
	ctx = context.WithValue(ctx, models.PrepareStepFnContextKey, params.PrepareStep)
	ctx = context.WithValue(ctx, models.ToolMiddlewaresContextKey, params.ToolMiddlewares)
	ctx = context.WithValue(ctx, models.StepMiddlewaresContextKey, params.StepMiddlewares)

	config := graph.InvocationConfig[agentState, agentStateDelta]{
		Checkpointer: resolveGraphCheckpointer(params.Checkpointer),
	}

	if params.Resume {
		g, ok := retrieveGraphForResume(params.Checkpointer, params.ThreadID)
		if !ok {
			g = createAgentGraph()
		}
		result, err := g.Resume(ctx, params.ThreadID, params.InterruptResults, config)
		if err != nil {
			return models.LanguageModelOutput{}, err
		}
		return resolveExecutionContextAsTextOutput(result.toolExecutionsArchive)
	}

	messages := resolveMessages(params)
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
				storeGraphForResume(params.Checkpointer, interrupts[0].ThreadID, g)
			}
		}
		return models.LanguageModelOutput{}, err
	}

	return resolveExecutionContextAsTextOutput(result.toolExecutionsArchive)
}
