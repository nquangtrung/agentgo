package agentgo

import (
	"testing"

	"github.com/nquangtrung/agentgo/providers/claude"
	"github.com/nquangtrung/agentgo/providers/deepseek"
	"github.com/nquangtrung/agentgo/providers/gemini"
	"github.com/nquangtrung/agentgo/providers/openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindSupportedModel(t *testing.T) {
	tests := []struct {
		modelName string
		expected  ModelType
		supported bool
	}{
		{"gpt-5-mini", MODEL_OPENAI, true},
		{"gemini-2.0-flash", MODEL_GEMINI, true},
		{"gemini-3.8-flash", MODEL_GEMINI, true},
		{"claude-sonnet-4", MODEL_CLAUDE, true},
		{"claude-opus-5-5", MODEL_CLAUDE, true},
		{"deepseek-flash", MODEL_DEEPSEEK, true},
		{"deepseek-v4-pro", MODEL_DEEPSEEK, true},
		{"deepseek-chat", MODEL_DEEPSEEK, true},
		{"mistral-large", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.modelName, func(t *testing.T) {
			modelType, supported := FindSupportedModel(tt.modelName)
			assert.Equal(t, tt.supported, supported)
			assert.Equal(t, tt.expected, modelType)
		})
	}
}

func TestLoadAPIKeyFromEnv(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai-key")
	t.Setenv("DEEPSEEK_API_KEY", "deepseek-key")
	t.Setenv("CLAUDE_API_KEY", "claude-key")
	t.Setenv("GEMINI_API_KEY", "gemini-key")

	apiKey, err := LoadAPIKeyFromEnv(MODEL_OPENAI)
	require.NoError(t, err)
	assert.Equal(t, "openai-key", apiKey)

	apiKey, err = LoadAPIKeyFromEnv(MODEL_DEEPSEEK)
	require.NoError(t, err)
	assert.Equal(t, "deepseek-key", apiKey)

	apiKey, err = LoadAPIKeyFromEnv(MODEL_CLAUDE)
	require.NoError(t, err)
	assert.Equal(t, "claude-key", apiKey)

	apiKey, err = LoadAPIKeyFromEnv(MODEL_GEMINI)
	require.NoError(t, err)
	assert.Equal(t, "gemini-key", apiKey)
}

func TestLoadAPIKeyFromEnvUnsupported(t *testing.T) {
	_, err := LoadAPIKeyFromEnv(ModelType("mistral"))
	assert.Error(t, err)
}

func TestCreateAgentProviderDeepSeek(t *testing.T) {
	provider, err := CreateAgentProvider(AgentProviderFactoryParams{
		APIKey:    "test-key",
		ModelName: "deepseek-flash",
	})

	require.NoError(t, err)
	deepSeekProvider, ok := provider.(deepseek.DeepSeekProvider)
	require.True(t, ok, "expected a deepseek provider, got %T", provider)
	assert.Equal(t, "deepseek-flash", deepSeekProvider.Context().ModelName)
}

func TestCreateAgentProviderOpenAI(t *testing.T) {
	provider, err := CreateAgentProvider(AgentProviderFactoryParams{
		APIKey:    "test-key",
		ModelName: "gpt-5-mini",
	})

	require.NoError(t, err)
	_, ok := provider.(openai.OpenAIProvider)
	assert.True(t, ok, "expected an openai provider, got %T", provider)
}

func TestCreateAgentProviderLoadsDeepSeekKeyFromEnv(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "deepseek-key")

	provider, err := CreateAgentProvider(AgentProviderFactoryParams{ModelName: "deepseek-flash"})

	require.NoError(t, err)
	_, ok := provider.(deepseek.DeepSeekProvider)
	assert.True(t, ok, "expected a deepseek provider, got %T", provider)
}

func TestCreateAgentProviderClaude(t *testing.T) {
	provider, err := CreateAgentProvider(AgentProviderFactoryParams{
		APIKey:    "test-key",
		ModelName: "claude-sonnet-4-6",
	})

	require.NoError(t, err)
	claudeProvider, ok := provider.(claude.ClaudeProvider)
	require.True(t, ok, "expected a claude provider, got %T", provider)
	assert.Equal(t, "claude-sonnet-4-6", claudeProvider.Context().ModelName)
}

func TestCreateAgentProviderGemini(t *testing.T) {
	provider, err := CreateAgentProvider(AgentProviderFactoryParams{
		APIKey:    "test-key",
		ModelName: "gemini-3.8-flash",
	})

	require.NoError(t, err)
	geminiProvider, ok := provider.(gemini.GeminiProvider)
	require.True(t, ok, "expected a gemini provider, got %T", provider)
	assert.Equal(t, "gemini-3.8-flash", geminiProvider.Context().ModelName)
}

func TestCreateAgentProviderLoadsClaudeAndGeminiKeysFromEnv(t *testing.T) {
	t.Setenv("CLAUDE_API_KEY", "claude-key")
	t.Setenv("GEMINI_API_KEY", "gemini-key")

	provider, err := CreateAgentProvider(AgentProviderFactoryParams{ModelName: "claude-sonnet-4-6"})
	require.NoError(t, err)
	_, ok := provider.(claude.ClaudeProvider)
	assert.True(t, ok, "expected a claude provider, got %T", provider)

	provider, err = CreateAgentProvider(AgentProviderFactoryParams{ModelName: "gemini-3.8-flash"})
	require.NoError(t, err)
	_, ok = provider.(gemini.GeminiProvider)
	assert.True(t, ok, "expected a gemini provider, got %T", provider)
}

func TestCreateAgentProviderUnsupportedModel(t *testing.T) {
	_, err := CreateAgentProvider(AgentProviderFactoryParams{
		APIKey:    "test-key",
		ModelName: "mistral-large",
	})

	assert.Error(t, err)
}

func TestCreateAgentProviderMissingKey(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")

	_, err := CreateAgentProvider(AgentProviderFactoryParams{ModelName: "deepseek-flash"})

	assert.Error(t, err)
}
