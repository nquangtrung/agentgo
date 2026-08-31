package fsm

import (
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
}
