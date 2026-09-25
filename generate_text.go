package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/models"
)

func GenerateText(ctx context.Context, params Params) (models.LanguageModelOutput, error) {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)
	execArchive := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	emitter := models.NewEmptyPartEmitter()

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, params.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, params.Tools)
	ctx = context.WithValue(ctx, models.StreamContextKey, false)
	ctx = context.WithValue(ctx, models.PartEmitterContextKey, emitter)
	ctx = context.WithValue(ctx, models.PrepareStepFnContextKey, params.PrepareStep)

	g := createGenerateTextGraph()
	config := graph.InvocationConfig[agentState, agentStateDelta]{}
	result, err := g.Invoke(ctx, agentState{
		toolExecutionsArchive: execArchive,
		messages:              messages,
	}, config)

	if err != nil {
		return models.LanguageModelOutput{}, err
	}

	return resolveExecutionContextAsTextOutput(result.toolExecutionsArchive)
}
