package agentgo

import (
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

type Params struct {
	Provider     providers.AgentProvider
	Prompt       string
	ModelName    string
	Messages     []models.Message
	Tools        []models.Tool
	MaxToolSteps int
}
