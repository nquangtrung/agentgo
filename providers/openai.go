package providers

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"trontria.com/agentgo/models"
)

type OpenAIProvider struct {
	AgentProviderImpl

	APIKey string
}

func (p OpenAIProvider) GenerateText(prompt string) (models.LanguageModelOutput, error) {
	client := openai.NewClient(
		option.WithAPIKey(p.APIKey),
	)

	resp, err := client.Responses.New(context.TODO(), responses.ResponseNewParams{
		Model: p.AgentProviderImpl.Context.ModelName,
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
	})
	if err != nil {
		return models.LanguageModelOutput{}, err
	}

	return models.LanguageModelOutput{
		Text: resp.OutputText(),
		Usage: models.LanguageModelUsage{
			OutputTokens:    int(resp.Usage.OutputTokens),
			InputTokens:     int(resp.Usage.InputTokens),
			CachedTokens:    int(resp.Usage.InputTokensDetails.CachedTokens) + int(resp.Usage.InputTokensDetails.CacheWriteTokens),
			ReasoningTokens: int(resp.Usage.OutputTokensDetails.ReasoningTokens),
		},
		ModelName: p.AgentProviderImpl.Context.ModelName,
	}, nil
}

func NewOpenAIProvider(apiKey, modelName string) OpenAIProvider {
	return OpenAIProvider{
		APIKey: apiKey,
		AgentProviderImpl: AgentProviderImpl{
			Context: models.NewLanguageModelContext(modelName),
		},
	}
}
