package agentgo

import (
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

func GenerateText(provider providers.AgentProvider, prompt string) (models.LanguageModelOutput, error) {
	return provider.GenerateText(prompt)
}
