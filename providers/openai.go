package providers

import (
	"context"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"
)

type OpenAIProvider struct {
	AgentProviderImpl

	APIKey string
}

func (p OpenAIProvider) GetInputFromParams(params AgentProviderPromptMessageParams) responses.ResponseNewParamsInputUnion {
	if len(params.Messages) > 0 {
		return p.ConvertMessageObjectToInput(params.Messages)
	}

	return responses.ResponseNewParamsInputUnion{OfString: openai.String(params.Prompt)}
}

func (p OpenAIProvider) ConvertMessageObjectToInput(messages []models.Message) responses.ResponseNewParamsInputUnion {
	var inputItems []responses.ResponseInputItemUnionParam = utils.Map(messages, func(message models.Message) responses.ResponseInputItemUnionParam {
		switch message.GetType() {
		case models.MessageRoleSystem:
			return responses.ResponseInputItemParamOfMessage(
				responses.ResponseInputMessageContentListParam{
					responses.ResponseInputContentParamOfInputText(message.GetContent().GetText()),
				},
				responses.EasyInputMessageRoleSystem,
			)
		case models.MessageRoleHuman:
			return responses.ResponseInputItemParamOfMessage(
				responses.ResponseInputMessageContentListParam{
					responses.ResponseInputContentParamOfInputText(message.GetContent().GetText()),
				},
				responses.EasyInputMessageRoleUser,
			)
		case models.MessageRoleAssistant:
			return responses.ResponseInputItemParamOfOutputMessage(
				[]responses.ResponseOutputMessageContentUnionParam{
					{OfOutputText: &responses.ResponseOutputTextParam{
						Text: message.GetContent().GetText(),
						// Annotations: []responses.ResponseOutputTextAnnotationUnionParam{},
					}},
				},
				"",
				responses.ResponseOutputMessageStatusCompleted,
			)
		default:
			return responses.ResponseInputItemParamOfMessage(
				responses.ResponseInputMessageContentListParam{
					responses.ResponseInputContentParamOfInputText(message.GetContent().GetText()),
				},
				responses.EasyInputMessageRoleUser,
			)
		}
	})

	return responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems}
}

func (p OpenAIProvider) GenerateText(params AgentProviderGenerateTextParams) (models.LanguageModelOutput, error) {
	client := openai.NewClient(
		option.WithAPIKey(p.APIKey),
	)

	resp, err := client.Responses.New(context.TODO(), responses.ResponseNewParams{
		Model: p.AgentProviderImpl.Context.ModelName,
		Input: p.GetInputFromParams(params.AgentProviderPromptMessageParams),
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

func (p OpenAIProvider) StreamText(params AgentProviderStreamTextParams, channel chan models.Part) {
	client := openai.NewClient(
		option.WithAPIKey(p.APIKey),
	)
	stream := client.Responses.NewStreaming(context.Background(), responses.ResponseNewParams{
		Model: p.AgentProviderImpl.Context.ModelName,
		Input: p.GetInputFromParams(params.AgentProviderPromptMessageParams),
	})

	channel <- models.NewStepStartPart(p.GetContext(), "streaming started")

	for stream.Next() {
		chunk := stream.Current()
		log.Println("Received chunk:", chunk.Type)
		switch chunk.Type {
		case "response.output_text.delta":
			channel <- models.NewTextPart(p.GetContext(), string(chunk.Delta))
		case "response.completed":
			channel <- models.NewStepEndPart(p.GetContext(), "stream completed", models.NewLanguageModelUsage(
				int(chunk.Response.Usage.OutputTokens),
				int(chunk.Response.Usage.InputTokens),
				int(chunk.Response.Usage.InputTokensDetails.CachedTokens)+int(chunk.Response.Usage.InputTokensDetails.CacheWriteTokens),
				int(chunk.Response.Usage.OutputTokensDetails.ReasoningTokens),
			))
		}
	}
}

func NewOpenAIProvider(apiKey, modelName string) OpenAIProvider {
	return OpenAIProvider{
		APIKey: apiKey,
		AgentProviderImpl: AgentProviderImpl{
			Context: models.NewLanguageModelContext(modelName),
		},
	}
}
