package fsm

import (
	"trontria.com/agentgo/models"
)

type AgentContext struct {
	ExecutionContext *models.ExecutionContext
	Messages         *[]models.Message
	TextGenerated    bool
}
