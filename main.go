package agentgo

import (
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

type GenerateTextParams struct {
	Provider providers.AgentProvider
	Prompt   string
}

func GenerateText(params GenerateTextParams) (models.LanguageModelOutput, error) {
	return params.Provider.GenerateText(params.Prompt)
}
