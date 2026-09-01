package fsm

import "trontria.com/agentgo/models"

type ToolChoice struct {
	Name string
}

type PrepareStepResult struct {
	// Override the tool choice for the current step. If nil, the default tool choice will be used.
	ToolChoice *ToolChoice

	// Override the messages for the current step. If nil, the default messages will be used.
	Messages *[]models.Message

	// Override the active tools for the current step. If nil, the default active tools will be used.
	ActiveTools *[]string
}

type PrepareStepFn func(step Step, ctx AgentContext) (PrepareStepResult, error)
