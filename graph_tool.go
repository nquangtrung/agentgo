package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/utils"
)

func executeTool(ctx context.Context, params any) (agentStateDelta, error) {
	toolCall := params.(models.ToolCall)
	if toolCall.NotFound {
		logger.Error("Tool not found", "toolCall", toolCall)
		return agentStateDelta{
			addError: &models.ToolNotFoundError{ToolName: toolCall.ToolName},
		}, nil
	}

	logger.Info("Executing tool", "tool", toolCall.Tool.Name(), "params", toolCall.Params)
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

	messages := state.messages
	if state.currentStep.prepareStepResult.Messages != nil {
		messages = *state.currentStep.prepareStepResult.Messages
	}

	tools := ctx.Value(models.ToolsContextKey).([]models.BaseTool)
	if state.currentStep.prepareStepResult.ActiveTools != nil {
		tools = utils.Filter(tools, func(t models.BaseTool) bool {
			return utils.Contains(
				*state.currentStep.prepareStepResult.ActiveTools,
				t.Name(),
				func(a, b string) bool { return a == b },
			)
		})
	}

	resolveOutput, err := provider.ResolveToolCall(
		ctx, providers.AgentProviderPromptMessageParams{Messages: messages},
		tools,
	)

	calls := []models.ToolCall{}
	for _, call := range resolveOutput.ToolCalls {
		tool, found := utils.Find(tools, func(t models.BaseTool) bool {
			return t.Name() == call.ToolName
		})
		call.NotFound = !found
		call.Tool = &tool
		calls = append(calls, call)
	}

	if err != nil {
		logger.Warn("Error resolving tool calls", "calls", calls, "usage", resolveOutput.Usage, "error", err)
		return agentStateDelta{}, err
	}

	logger.Info("Resolved tool calls", "calls", calls, "usage", resolveOutput.Usage)
	return agentStateDelta{
		from:           RESOLVE_TOOL,
		availableTools: calls,
	}, nil
}
