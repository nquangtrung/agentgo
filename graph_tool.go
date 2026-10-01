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

	toolMiddlewares := ctx.Value(models.ToolMiddlewaresContextKey).([]ToolMiddleware)

	// Run Before hooks
	for _, mw := range toolMiddlewares {
		if mw.Before != nil {
			err := mw.Before(ctx, ToolMiddlewareContext{
				ToolName: toolCall.ToolName,
				Params:   toolCall.Params,
			})
			if err != nil {
				logger.Warn("Tool middleware Before hook error", "tool", toolCall.Tool.Name(), "error", err)
				// If it's an interrupt error, propagate it to pause execution
				if isInterruptError(err) {
					return agentStateDelta{}, err
				}
				// Otherwise, include the error in the tool result and continue
				toolResult := models.ToolExecuteOutput{
					Error:    err,
					ToolCall: &toolCall,
				}
				return agentStateDelta{
					from:              EXECUTE_TOOL,
					archiveToolResult: &toolResult,
				}, nil
			}
		}
	}

	logger.Info("Executing tool", "tool", toolCall.Tool.Name(), "params", toolCall.Params)
	tool := toolCall.Tool
	toolResult := tool.Execute(models.ToolExecuteParams{
		Input: toolCall.Params,
	})
	toolResult.ToolCall = &toolCall

	logger.Debug("Tool executed", "tool", tool.Name(), "result", toolResult)

	// Run After hooks only on success (no error from Execute)
	if toolResult.Error == nil {
		for _, mw := range toolMiddlewares {
			if mw.After != nil {
				err := mw.After(ctx, ToolMiddlewareContext{
					ToolName: toolCall.ToolName,
					Params:   toolCall.Params,
					Result:   &toolResult,
				})
				if err != nil {
					logger.Warn("Tool middleware After hook error", "tool", toolCall.Tool.Name(), "error", err)
					// If it's an interrupt error, propagate it
					if isInterruptError(err) {
						return agentStateDelta{}, err
					}
					// Otherwise, include the validation error in the result
					toolResult.Error = err
					break
				}
			}
		}
	}

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

	// Record the model's own requests in the conversation before any results
	// arrive. Providers with structured tool protocols require a result to
	// reference the call that produced it, so the call turn has to precede the
	// result turns that archiveToolResult appends.
	var appendMessages []models.Message
	if len(calls) > 0 {
		appendMessages = append(appendMessages, models.NewAssistantToolCallsMessage(calls, resolveOutput.Text))
	}

	return agentStateDelta{
		from:           RESOLVE_TOOL,
		availableTools: calls,
		appendMessages: appendMessages,
	}, nil
}
