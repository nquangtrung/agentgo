package providers

import (
	"strings"

	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"
)

type AgentProvider interface {
	GetContext() models.LanguageModelContext
	GenerateText(prompt string) (models.LanguageModelOutput, error)
	StreamText(prompt string) chan models.Part
}

type AgentProviderImpl struct {
	Context models.LanguageModelContext
}

func (p AgentProviderImpl) GetContext() models.LanguageModelContext {
	return p.Context
}

type AgentProviderFactoryParams struct {
	APIKey    string
	ModelName string
}

type ModelType string

const (
	MODEL_OPENAI ModelType = "openai"
	MODEL_GEMINI ModelType = "gemini"
	MODEL_CLAUDE ModelType = "claude"
)

func FindSupportedModel(modelName string) (ModelType, bool) {
	switch {
	case strings.HasPrefix(modelName, "gpt"):
		return MODEL_OPENAI, true
	case strings.HasPrefix(modelName, "gemini"):
		return MODEL_GEMINI, true
	case strings.HasPrefix(modelName, "claude"):
		return MODEL_CLAUDE, true
	default:
		return "", false
	}
}

func LoadAPIKeyFromEnv(modelType ModelType) (string, error) {
	switch modelType {
	case MODEL_OPENAI:
		return utils.GetEnvVar("OPENAI_API_KEY"), nil
	case MODEL_GEMINI:
		return utils.GetEnvVar("GEMINI_API_KEY"), nil
	case MODEL_CLAUDE:
		return utils.GetEnvVar("CLAUDE_API_KEY"), nil
	default:
		return "", &models.UnsupportedModelError{ModelName: string(modelType)}
	}
}

func CreateAgentProvider(params AgentProviderFactoryParams) (AgentProvider, error) {
	modelType, supported := FindSupportedModel(params.ModelName)
	if !supported {
		return nil, &models.UnsupportedModelError{ModelName: "api key not found for model: " + params.ModelName}
	}

	if params.APIKey == "" {
		apiKey, err := LoadAPIKeyFromEnv(modelType)
		if err != nil {
			return nil, &models.UnsupportedModelError{ModelName: "api key not found for model: " + params.ModelName}
		}
		params.APIKey = apiKey
	}

	if params.APIKey == "" {
		return nil, &models.UnsupportedModelError{ModelName: "api key not found for model: " + params.ModelName}
	}

	switch modelType {
	case MODEL_OPENAI:
		return NewOpenAIProvider(params.APIKey, params.ModelName), nil
	case MODEL_GEMINI:
		// return NewGeminiProvider(params.APIKey, params.ModelName), nil
	case MODEL_CLAUDE:
		// return NewClaudeProvider(params.APIKey, params.ModelName), nil
	default:
		return nil, &models.UnsupportedModelError{ModelName: params.ModelName}
	}

	return nil, &models.UnsupportedModelError{ModelName: params.ModelName}
}
