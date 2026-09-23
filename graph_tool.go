package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

func executeTool(ctx context.Context, params any) (agentStateDelta, error) {
	toolCall := params.(models.ToolCall)
	tool := toolCall.Tool
	toolResult := tool.Execute(models.ToolExecuteParams{
		Input: toolCall.Params,
	})

	logger.Debug("Tool executed", "tool", tool.Name(), "result", toolResult)

	return agentStateDelta{
		from:              EXECUTE_TOOL,
		archiveToolResult: &toolResult,
	}, nil
}

func resolveTool(ctx context.Context, state agentState) (agentStateDelta, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)

	// TODO Handle when prepare step return only a subset
	tools := ctx.Value(models.ToolsContextKey).([]models.BaseTool)
	resolveOutput, err := provider.ResolveToolCall(
		ctx, providers.AgentProviderPromptMessageParams{Messages: *state.messages},
		tools,
	)

	if err != nil {
		// TODO Handle error
		return agentStateDelta{}, err
	}

	return agentStateDelta{
		from:           RESOLVE_TOOL,
		availableTools: resolveOutput.ToolCalls,
	}, nil
}
