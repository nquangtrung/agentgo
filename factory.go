package agentgo

import (
	"strings"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/providers/claude"
	"github.com/nquangtrung/agentgo/providers/deepseek"
	"github.com/nquangtrung/agentgo/providers/gemini"
	"github.com/nquangtrung/agentgo/providers/openai"
	"github.com/nquangtrung/agentgo/utils"
)

type AgentProviderFactoryParams struct {
	APIKey    string
	ModelName string
}

type ModelType string

const (
	MODEL_OPENAI   ModelType = "openai"
	MODEL_GEMINI   ModelType = "gemini"
	MODEL_CLAUDE   ModelType = "claude"
	MODEL_DEEPSEEK ModelType = "deepseek"
)

func FindSupportedModel(modelName string) (ModelType, bool) {
	switch {
	case strings.HasPrefix(modelName, "gpt"):
		return MODEL_OPENAI, true
	case strings.HasPrefix(modelName, "gemini"):
		return MODEL_GEMINI, true
	case strings.HasPrefix(modelName, "claude"):
		return MODEL_CLAUDE, true
	case strings.HasPrefix(modelName, "deepseek"):
		return MODEL_DEEPSEEK, true
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
	case MODEL_DEEPSEEK:
		return utils.GetEnvVar("DEEPSEEK_API_KEY"), nil
	default:
		return "", &models.UnsupportedModelError{ModelName: string(modelType)}
	}
}

func CreateAgentProvider(params AgentProviderFactoryParams) (providers.AgentProvider, error) {
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
		return openai.NewOpenAIProvider(params.APIKey, params.ModelName), nil
	case MODEL_DEEPSEEK:
		return deepseek.NewDeepSeekProvider(params.APIKey, params.ModelName), nil
	case MODEL_GEMINI:
		return gemini.NewGeminiProvider(params.APIKey, params.ModelName), nil
	case MODEL_CLAUDE:
		return claude.NewClaudeProvider(params.APIKey, params.ModelName), nil
	default:
		return nil, &models.UnsupportedModelError{ModelName: params.ModelName}
	}
}
