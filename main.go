package agentgo

import (
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

type GenerateTextParams struct {
	Provider  providers.AgentProvider
	Prompt    string
	ModelName string
	Messages  []models.Message
}

func GenerateText(params GenerateTextParams) (models.LanguageModelOutput, error) {
	if params.ModelName != "" {
		var provider, err = providers.CreateAgentProvider(providers.AgentProviderFactoryParams{
			ModelName: params.ModelName,
		})
		if err != nil {
			return models.LanguageModelOutput{}, err
		}
		params.Provider = provider
	}

	if params.Provider == nil {
		return models.LanguageModelOutput{}, &models.UnsupportedModelError{ModelName: "nil provider"}
	}

	return params.Provider.GenerateText(providers.AgentProviderGenerateTextParams{
		AgentProviderPromptMessageParams: providers.AgentProviderPromptMessageParams{
			Prompt:   params.Prompt,
			Messages: params.Messages,
		},
	})
}

type StreamTextParams struct {
	Provider  providers.AgentProvider
	Prompt    string
	ModelName string
	Messages  []models.Message
}

func StreamText(params StreamTextParams) models.LanguageModelStreamOutput {
	if params.ModelName != "" {
		var provider, err = providers.CreateAgentProvider(providers.AgentProviderFactoryParams{
			ModelName: params.ModelName,
		})
		if err != nil {
			panic(err)
		}
		params.Provider = provider
	}

	if params.Provider == nil {
		panic("nil provider")
	}

	channel := make(chan models.Part)
	go func() {
		defer close(channel)
		params.Provider.StreamText(providers.AgentProviderStreamTextParams{
			AgentProviderPromptMessageParams: providers.AgentProviderPromptMessageParams{
				Prompt:   params.Prompt,
				Messages: params.Messages,
			},
		}, channel)
	}()

	return models.NewLanguageModelStreamOutput(channel, params.Provider.GetContext().ModelName)
}
