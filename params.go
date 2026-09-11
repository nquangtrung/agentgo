package agentgo

import (
	"github.com/nquangtrung/agentgo/fsm"
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

type Params struct {
	Provider      providers.AgentProvider
	Prompt        string
	ModelName     string
	Messages      []models.Message
	Tools         []models.BaseTool
	EndConditions []models.EndCondition
	PrepareStep   fsm.PrepareStepFn
}
