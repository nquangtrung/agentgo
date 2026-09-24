package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/utils"
)

func executeTool(ctx context.Context, params any) (agentStateDelta, error) {
	toolCall := params.(models.ToolCall)
	if toolCall.Tool == nil {
		logger.Error("Tool is nil", "toolCall", toolCall)
		return agentStateDelta{}, nil
	}
	logger.Debug("Executing tool", "tool", toolCall.Tool.Name(), "params", toolCall.Params)
	tool := toolCall.Tool
	toolResult := tool.Execute(models.ToolExecuteParams{
		Input: toolCall.Params,
	})
	toolResult.ToolCall = &toolCall

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
		ctx, providers.AgentProviderPromptMessageParams{Messages: state.messages},
		tools,
	)

	calls := []models.ToolCall{}
	for _, call := range resolveOutput.ToolCalls {
		tool, _ := utils.Find(tools, func(t models.BaseTool) bool {
			return t.Name() == call.ToolName
		})
		logger.Info("Resolved tool call", "toolCall", call, "tool", tool)
		call.Tool = &tool
		calls = append(calls, call)
	}

	if err != nil {
		// TODO Handle error
		return agentStateDelta{}, err
	}

	return agentStateDelta{
		from:           RESOLVE_TOOL,
		availableTools: calls,
	}, nil
}
