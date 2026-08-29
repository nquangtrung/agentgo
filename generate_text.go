package agentgo

import (
	"context"

	"trontria.com/agentgo/fsm"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"
)

func GenerateText(ctx context.Context, params Params) (models.LanguageModelOutput, error) {
	provider := mustResolveProviderFromParams(params)
	messages := resolveMessages(params)
	execContext := models.NewExecutionContextFromLanguageModelContext(provider.Context())

	accumulator := func(acc *models.ExecutionContext, item *models.ToolExecuteOutput) {
		accumulateToolCallResult(acc, item, &messages)
	}
	runner := utils.NewRunner(nil, accumulator)
	machine := fsm.New[fsm.AgentContext]()

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)
	ctx = context.WithValue(ctx, models.RunnerContextKey, runner)
	ctx = context.WithValue(ctx, models.MachineContextKey, machine)
	ctx = context.WithValue(ctx, models.EndConditionsContextKey, params.EndConditions)
	ctx = context.WithValue(ctx, models.ToolsContextKey, params.Tools)

	agentContext := fsm.AgentContext{
		ExecutionContext: execContext,
		Messages:         &messages,
	}
	machine.Run(ctx, &fsm.PredicateState{}, &agentContext)

	return resolveExecutionContextAsTextOutput(execContext)
}
