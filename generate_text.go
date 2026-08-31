package agentgo

import (
	"context"

	"trontria.com/agentgo/fsm"
	"trontria.com/agentgo/models"
)

func GenerateText(ctx context.Context, params Params) (models.LanguageModelOutput, error) {
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
