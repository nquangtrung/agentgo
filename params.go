package agentgo

import (
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

type Params struct {
	Provider         providers.AgentProvider
	Prompt           string
	ModelName        string
	Messages         []models.Message
	Tools            []models.BaseTool
	EndConditions    []models.EndCondition
	PrepareStep      PrepareStepFn
	ToolMiddlewares  []ToolMiddleware
	StepMiddlewares  []StepMiddleware
	Checkpointer     Checkpointer
	// Resume fields — set these to resume an execution paused by a HITL interrupt.
	Resume           bool
	ThreadID         string
	InterruptResults map[string]any
}
