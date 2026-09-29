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

// ObjectParams is the parameter set for GenerateObject and StreamObject.
// It mirrors Params (flat, not embedded, so composite literals stay ergonomic)
// and adds the schema plus the retry bound for the repair loop.
type ObjectParams[T any] struct {
	Provider        providers.AgentProvider
	Prompt          string
	ModelName       string
	Messages        []models.Message
	Tools           []models.BaseTool
	EndConditions   []models.EndCondition
	PrepareStep     PrepareStepFn
	ToolMiddlewares []ToolMiddleware
	StepMiddlewares []StepMiddleware
	Checkpointer    Checkpointer
	// Resume fields — set these to resume an execution paused by a HITL interrupt.
	Resume           bool
	ThreadID         string
	InterruptResults map[string]any

	// Schema describes the object the model must produce. Required.
	Schema models.Schema[T]
	// MaxObjectRetries bounds the total number of object generation attempts
	// (including the first). Defaults to 2.
	MaxObjectRetries int
}

// ObjectResult is the outcome of a successful object generation.
type ObjectResult[T any] struct {
	Object    T
	Raw       string
	Usage     models.LanguageModelUsage
	ModelName string
	Context   *models.ToolExecutionsArchive
}

// baseParams converts ObjectParams to the shared Params used by the graph.
func (p ObjectParams[T]) baseParams() Params {
	return Params{
		Provider:         p.Provider,
		Prompt:           p.Prompt,
		ModelName:        p.ModelName,
		Messages:         p.Messages,
		Tools:            p.Tools,
		EndConditions:    p.EndConditions,
		PrepareStep:      p.PrepareStep,
		ToolMiddlewares:  p.ToolMiddlewares,
		StepMiddlewares:  p.StepMiddlewares,
		Checkpointer:     p.Checkpointer,
		Resume:           p.Resume,
		ThreadID:         p.ThreadID,
		InterruptResults: p.InterruptResults,
	}
}
