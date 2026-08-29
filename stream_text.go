package agentgo

import (
	"context"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
)

func StreamText(ctx context.Context, params Params) models.LanguageModelStreamOutput {
	provider := mustResolveProviderFromParams(params)

	ctx = context.WithValue(ctx, models.ProviderContextKey, provider)

	channel := make(chan models.Part)
	go func() {
		defer close(channel)
		if len(params.EndConditions) == 0 || len(params.Tools) == 0 {
			messages := resolveMessages(params)
			provider.StreamText(ctx, providers.AgentProviderPromptMessageParams{Messages: messages}, channel)
			return
		}

		// doStreamLoop(ctx, params, channel)
	}()

	return models.NewLanguageModelStreamOutput(channel, params.Provider.Context().ModelName)
}
