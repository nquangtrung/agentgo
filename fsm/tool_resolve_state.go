package fsm

import (
	"context"
	"log"
	"sync"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
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
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)

	messages := fsmCtx.ResolveCurrentStepMessages(ctx)

	tools := fsmCtx.ResolveCurrentStepActiveTools(ctx)
	toolCalls, err := provider.ResolveToolCall(ctx, providers.AgentProviderPromptMessageParams{Messages: messages}, tools)
	if err != nil {
		return &RetryState{
			OriginalState: s,
		}, nil
	}

	if len(toolCalls) == 0 {
		log.Printf("No tool calls resolved")
		return &PrepareTextGenerationState{}, nil
	}

	var wg sync.WaitGroup
	for _, toolCall := range toolCalls {
		wg.Go(func() {
			log.Printf("Executing tool %s", toolCall.ToolName)

			emitter.Emit(models.NewToolStartPart(
				provider.Context(),
				toolCall.ToolName,
			))
			result, err := executeToolCall(ctx, toolCall, tools)

			switch {
			case err != nil:
				// This error is before request to the provider, so usage is 0.
				log.Printf("Error executing tool %s: %v", toolCall.ToolName, err)
				emitter.Emit(models.NewToolErrorPart(
					provider.Context(),
					toolCall.ToolName,
					models.NewLanguageModelUsage(0, 0, 0, 0),
					err,
				))
			case result.Error != nil:
				emitter.Emit(models.NewToolErrorPart(
					provider.Context(),
					toolCall.ToolName,
					result.Usage,
					result.Error,
				))
				models.AccumulateToolCallResult(fsmCtx.ToolExecutionsArchive, result, fsmCtx.Messages)
				fsmCtx.CurrentStep.Usage = models.AccumulateUsage(fsmCtx.CurrentStep.Usage, result.Usage)
			default:
				emitter.Emit(models.NewToolResultPart(
					provider.Context(),
					toolCall.ToolName,
					*result,
				))
				models.AccumulateToolCallResult(fsmCtx.ToolExecutionsArchive, result, fsmCtx.Messages)
				fsmCtx.CurrentStep.Usage = models.AccumulateUsage(fsmCtx.CurrentStep.Usage, result.Usage)
			}
		})
	}
	wg.Wait()
	return &StepEndState{}, nil
}
