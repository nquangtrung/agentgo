package providers

import (
	"trontria.com/agentgo/models"
)

type AgentProvider interface {
	GenerateText(prompt string) (models.LanguageModelOutput, error)
}
