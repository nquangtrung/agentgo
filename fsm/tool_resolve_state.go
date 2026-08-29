package fsm

import (
	"context"
	"log"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

type ToolResolveState struct {
}

func resolveToolFromToolCall(toolCall models.ToolCall, tools []models.BaseTool) (models.Tool, error) {
	for _, tool := range tools {
		if tool.Name() == toolCall.ToolName {
			return tool, nil
		}
	}
	return nil, &models.ToolNotFoundError{ToolName: toolCall.ToolName}
}

func executeToolCall(_ context.Context, toolCall models.ToolCall, tools []models.BaseTool) (*models.ToolExecuteOutput, error) {
	tool, err := resolveToolFromToolCall(toolCall, tools)
	if err != nil {
		return nil, err
	}
	log.Printf("Resolved tool [%s]", tool.Name())
	toolResult := tool.Execute(models.ToolExecuteParams{
		Input: toolCall.Params,
	})
	log.Printf("Executed tool %s with result: %v", tool.Name(), toolResult)
	toolResult.ToolCall = &toolCall
	return &toolResult, nil
}

func (s *ToolResolveState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	tools := ctx.Value(models.ToolsContextKey).([]models.BaseTool)
	runner := ctx.Value(models.RunnerContextKey).(*utils.Runner[*models.ToolExecuteOutput, models.ExecutionContext])

	messages := fsmCtx.Messages
	execCtx := fsmCtx.ExecutionContext

	toolCalls, err := provider.ResolveToolCall(ctx, providers.AgentProviderPromptMessageParams{Messages: *messages}, tools)
	if err != nil {
		return &PredicateState{}, err
	}

	if len(toolCalls) == 0 {
		log.Printf("No tool calls resolved")
		return &TextGenerationState{}, nil
	}

	utils.Each(toolCalls, func(toolCall models.ToolCall) {
		log.Printf("Executing tool %s", toolCall.ToolName)
		runner.Execute(
			"tool",
			func(execCtx *models.ExecutionContext) (*models.ToolExecuteOutput, error) {
				result, err := executeToolCall(ctx, toolCall, tools)
				return result, err
			},
			execCtx,
		)
	})

	return &PredicateState{}, nil
}
