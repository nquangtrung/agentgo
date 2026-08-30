package fsm

import (
	"trontria.com/agentgo/models"
)

type Step struct {
	StepIndex int
	Usage     models.LanguageModelUsage
}

type AgentContext struct {
	ToolExecutionsArchive *models.ToolExecutionsArchive
	Messages              *[]models.Message
	TextGenerated         bool

	Steps       []Step
	CurrentStep Step

	TotalUsage models.LanguageModelUsage
}
