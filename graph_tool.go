package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

func executeTool(ctx context.Context, params any) (agentState, error) {
	toolCall := params.(models.ToolCall)
	tool := toolCall.Tool
	toolResult := tool.Execute(models.ToolExecuteParams{
		Input: toolCall.Params,
	})

	return agentState{
		currentStep: step{
			toolResults: []models.ToolExecuteOutput{toolResult},
		},
	}, nil
}

func resolveTool(ctx context.Context, state agentState) (agentState, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)

	// TODO Handle when prepare step return only a subset
	tools := ctx.Value(models.ToolsContextKey).([]models.BaseTool)
	resolveOutput, err := provider.ResolveToolCall(
		ctx, providers.AgentProviderPromptMessageParams{Messages: *state.messages},
		tools,
	)

	if err != nil {
		// TODO Handle error
		return state, err
	}

	currentStep := state.currentStep
	currentStep.tools = resolveOutput.ToolCalls
	return agentState{
		currentStep: currentStep,
	}, nil
}
