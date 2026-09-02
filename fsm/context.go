package fsm

import (
	"context"
	"time"

	"trontria.com/agentgo/models"
)

type Step struct {
	StepIndex         int
	Usage             models.LanguageModelUsage
	PrepareStepResult PrepareStepResult
}

type AgentContext struct {
	ToolExecutionsArchive *models.ToolExecutionsArchive
	Messages              *[]models.Message
	TextGenerated         bool

	Steps       []Step
	CurrentStep Step

	TotalUsage models.LanguageModelUsage

	RetryCount    int
	LastRetryWait time.Duration
}

func (ac *AgentContext) ResolveCurrentStepMessages(_ context.Context) []models.Message {
	if ac.CurrentStep.PrepareStepResult.Messages != nil {
		return *ac.CurrentStep.PrepareStepResult.Messages
	}
	return *ac.Messages
}

func (ac *AgentContext) ResolveCurrentStepActiveTools(ctx context.Context) []models.BaseTool {
	tools := ctx.Value(models.ToolsContextKey).([]models.BaseTool)
	switch {
	case ac.CurrentStep.PrepareStepResult.ActiveTools != nil:
		activeToolNames := *ac.CurrentStep.PrepareStepResult.ActiveTools
		filteredTools := []models.BaseTool{}
		for _, tool := range tools {
			for _, activeToolName := range activeToolNames {
				if tool.Name() == activeToolName {
					filteredTools = append(filteredTools, tool)
					break
				}
			}
		}
		return filteredTools
	case ac.CurrentStep.PrepareStepResult.ToolChoice != nil:
		toolChoice := ac.CurrentStep.PrepareStepResult.ToolChoice
		if toolChoice.Name != "" {
			for _, tool := range tools {
				if tool.Name() == toolChoice.Name {
					return []models.BaseTool{tool}
				}
			}
		}
		return tools
	default:
		return tools
	}
}
