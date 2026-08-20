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
	response openAIResponsesService
}

func (p OpenAIProvider) GetInputFromParams(params providers.AgentProviderPromptMessageParams) responses.ResponseNewParamsInputUnion {
	if len(params.Messages) == 0 {
		panic("no messages provided to GetInputFromParams")
	}

	return p.ConvertMessageObjectToInput(params.Messages)
}

func (p OpenAIProvider) ConvertMessageObjectToInput(messages []models.Message) responses.ResponseNewParamsInputUnion {
	var inputItems []responses.ResponseInputItemUnionParam = utils.Map(
		messages,
		func(message models.Message) responses.ResponseInputItemUnionParam {
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
		},
	)

	return responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems}
}

func (p OpenAIProvider) GenerateText(params providers.AgentProviderPromptMessageParams) (models.LanguageModelOutput, error) {
	resp, err := p.response.New(context.Background(), responses.ResponseNewParams{
		Model: p.BaseAgentProvider.Context().ModelName,
		Input: p.GetInputFromParams(params),
	})
	if err != nil {
		return models.LanguageModelOutput{}, err
	}

	return models.LanguageModelOutput{
		Text: resp.OutputText(),
		Usage: models.LanguageModelUsage{
			OutputTokens:    int64(resp.Usage.OutputTokens),
			InputTokens:     int64(resp.Usage.InputTokens),
			CachedTokens:    int64(resp.Usage.InputTokensDetails.CachedTokens) + int64(resp.Usage.InputTokensDetails.CacheWriteTokens),
			ReasoningTokens: int64(resp.Usage.OutputTokensDetails.ReasoningTokens),
		},
		ModelName: p.BaseAgentProvider.Context().ModelName,
	}, nil
}

func (p OpenAIProvider) StreamText(params providers.AgentProviderPromptMessageParams, channel chan models.Part) {
	stream := p.response.NewStreaming(context.Background(), responses.ResponseNewParams{
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
				int64(chunk.Response.Usage.OutputTokens),
				int64(chunk.Response.Usage.InputTokens),
				int64(chunk.Response.Usage.InputTokensDetails.CachedTokens)+int64(chunk.Response.Usage.InputTokensDetails.CacheWriteTokens),
				int64(chunk.Response.Usage.OutputTokensDetails.ReasoningTokens),
			))
		}
	}
}

func (p OpenAIProvider) ResolveToolCall(params providers.AgentProviderPromptMessageParams, toolParams []models.Tool) *models.ToolCall {
	return nil
}

func NewOpenAIProvider(apiKey, modelName string) OpenAIProvider {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	return newOpenAIProviderWithClient(
		modelName,
		&client.Responses,
	)
}

func newOpenAIProviderWithClient(modelName string, responses openAIResponsesService) OpenAIProvider {
	return OpenAIProvider{
		response: responses,
		BaseAgentProvider: providers.NewBaseAgentProvider(
			models.LanguageModelContext{ModelName: modelName},
		),
	}
}
