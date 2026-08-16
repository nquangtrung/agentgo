package openai

import (
	"context"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/providers"
	"trontria.com/agentgo/utils"
)

type OpenAIProvider struct {
	providers.BaseAgentProvider

	APIKey string
}

func (p OpenAIProvider) GetInputFromParams(params providers.AgentProviderPromptMessageParams) responses.ResponseNewParamsInputUnion {
	if len(params.Messages) == 0 {
		panic("no messages provided to GetInputFromParams")
	}

	return p.ConvertMessageObjectToInput(params.Messages)
}

func (p OpenAIProvider) ConvertMessageObjectToInput(messages []models.Message) responses.ResponseNewParamsInputUnion {
	var inputItems []responses.ResponseInputItemUnionParam = utils.Map(messages, func(message models.Message) responses.ResponseInputItemUnionParam {
		switch message.Type() {
		case models.MessageRoleSystem:
			return responses.ResponseInputItemParamOfMessage(
				responses.ResponseInputMessageContentListParam{
					responses.ResponseInputContentParamOfInputText(message.Content().Text()),
				},
				responses.EasyInputMessageRoleSystem,
			)
		case models.MessageRoleHuman:
			return responses.ResponseInputItemParamOfMessage(
				responses.ResponseInputMessageContentListParam{
					responses.ResponseInputContentParamOfInputText(message.Content().Text()),
				},
				responses.EasyInputMessageRoleUser,
			)
		case models.MessageRoleAssistant:
			return responses.ResponseInputItemParamOfOutputMessage(
				[]responses.ResponseOutputMessageContentUnionParam{
					{OfOutputText: &responses.ResponseOutputTextParam{
						Text: message.Content().Text(),
						// Annotations: []responses.ResponseOutputTextAnnotationUnionParam{},
					}},
				},
				"",
				responses.ResponseOutputMessageStatusCompleted,
			)
		default:
			return responses.ResponseInputItemParamOfMessage(
				responses.ResponseInputMessageContentListParam{
					responses.ResponseInputContentParamOfInputText(message.Content().Text()),
				},
				responses.EasyInputMessageRoleUser,
			)
		}
	})

	return responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems}
}

func (p OpenAIProvider) GenerateText(params providers.AgentProviderPromptMessageParams) (models.LanguageModelOutput, error) {
	client := openai.NewClient(
		option.WithAPIKey(p.APIKey),
	)

	resp, err := client.Responses.New(context.TODO(), responses.ResponseNewParams{
		Model: p.BaseAgentProvider.Context().ModelName,
		Input: p.GetInputFromParams(params),
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
		ModelName: p.BaseAgentProvider.Context().ModelName,
	}, nil
}

func (p OpenAIProvider) StreamText(params providers.AgentProviderPromptMessageParams, channel chan models.Part) {
	client := openai.NewClient(
		option.WithAPIKey(p.APIKey),
	)
	stream := client.Responses.NewStreaming(context.Background(), responses.ResponseNewParams{
		Model: p.BaseAgentProvider.Context().ModelName,
		Input: p.GetInputFromParams(params),
	})

	channel <- models.NewStepStartPart(p.Context(), "streaming started")

	for stream.Next() {
		chunk := stream.Current()
		log.Println("Received chunk:", chunk.Type)
		switch chunk.Type {
		case "response.output_text.delta":
			channel <- models.NewTextPart(p.Context(), string(chunk.Delta))
		case "response.completed":
			channel <- models.NewStepEndPart(p.Context(), "stream completed", models.NewLanguageModelUsage(
				int(chunk.Response.Usage.OutputTokens),
				int(chunk.Response.Usage.InputTokens),
				int(chunk.Response.Usage.InputTokensDetails.CachedTokens)+int(chunk.Response.Usage.InputTokensDetails.CacheWriteTokens),
				int(chunk.Response.Usage.OutputTokensDetails.ReasoningTokens),
			))
		}
	}
}

func (p OpenAIProvider) ResolveToolCall(params providers.AgentProviderPromptMessageParams, toolParams []models.Tool) *models.ToolCall {
	return nil
}

func NewOpenAIProvider(apiKey, modelName string) OpenAIProvider {
	return OpenAIProvider{
		APIKey: apiKey,
		BaseAgentProvider: providers.NewBaseAgentProvider(
			models.LanguageModelContext{ModelName: modelName},
		),
	}
}
