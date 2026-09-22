package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/fsm"
	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/models"
)

func GenerateText(ctx context.Context, params Params) (models.LanguageModelOutput, error) {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)
	execContext := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	emitter := models.NewEmptyPartEmitter()

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, params.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, params.Tools)
	ctx = context.WithValue(ctx, models.StreamContextKey, false)
	ctx = context.WithValue(ctx, models.PartEmitterContextKey, emitter)
	ctx = context.WithValue(ctx, models.PrepareStepFnContextKey, params.PrepareStep)

	g := createGenerateTextGraph()

	config := graph.InvocationConfig[agentState]{}
	result, err := g.Invoke(ctx, agentState{
		ToolExecutionsArchive: execContext,
		Messages:              &messages,
	}, config)

	if err != nil {
		return models.LanguageModelOutput{}, err
	}

	return resolveExecutionContextAsTextOutput(result.ToolExecutionsArchive)
}

func GenerateText_deprecated(ctx context.Context, params Params) (models.LanguageModelOutput, error) {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)
	execContext := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	machine := fsm.New[fsm.AgentContext]()
	emitter := models.NewEmptyPartEmitter()

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.MachineContextKey, machine)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, params.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, params.Tools)
	ctx = context.WithValue(ctx, models.StreamContextKey, false)
	ctx = context.WithValue(ctx, models.PartEmitterContextKey, emitter)
	ctx = context.WithValue(ctx, models.PrepareStepFnContextKey, params.PrepareStep)

	agentContext := fsm.AgentContext{
		ToolExecutionsArchive: execContext,
		Messages:              &messages,
	}
	machine.Run(ctx, &fsm.StartState{}, &agentContext)

	return resolveExecutionContextAsTextOutput(execContext)
}
