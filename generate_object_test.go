package agentgo

import (
	"context"
	"testing"

	"github.com/nquangtrung/agentgo/mocks"
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testRecipe struct {
	Name     string `json:"name"`
	Servings int    `json:"servings"`
}

func testRecipeSchema() models.Schema[testRecipe] {
	return models.NewSchema[testRecipe](models.NewSchemaParams[testRecipe]{
		Name:        "recipe",
		Description: "A recipe",
		JSONSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":     map[string]any{"type": "string"},
				"servings": map[string]any{"type": "integer"},
			},
			"required": []string{"name", "servings"},
		},
		Validate: func(r testRecipe) error {
			if r.Name == "" {
				return assert.AnError
			}
			return nil
		},
	})
}

func TestGenerateObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Generate a pancake recipe"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := ObjectParams[testRecipe]{
		Prompt:  prompt,
		Provider: mockProvider,
		Schema:  testRecipeSchema(),
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
			if len(p.Messages) != 1 || p.Messages[0].Content().Text() != prompt {
				return false
			}
			return p.ResponseFormat != nil && p.ResponseFormat.Name == "recipe"
		}),
	).Return(models.LanguageModelOutput{
		Text: `{"name":"pancakes","servings":4}`,
		Usage: models.LanguageModelUsage{
			OutputTokens: 123,
			InputTokens:  245,
			InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
				CachedTokens: 21,
			},
			OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
				ReasoningTokens: 12,
			},
			TotalTokens: 368,
		},
		ModelName: modelName,
	}, nil)

	result, err := GenerateObject(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, modelName, result.ModelName)
	assert.Equal(t, "pancakes", result.Object.Name)
	assert.Equal(t, 4, result.Object.Servings)
	assert.Equal(t, `{"name":"pancakes","servings":4}`, result.Raw)
	assert.Equal(t, int64(245), result.Usage.InputTokens)
	assert.Equal(t, int64(123), result.Usage.OutputTokens)
	assert.Equal(t, int64(21), result.Usage.InputTokensDetails.CachedTokens)
	assert.Equal(t, int64(12), result.Usage.OutputTokensDetails.ReasoningTokens)
	assert.Equal(t, int64(368), result.Usage.TotalTokens)
	assert.NotNil(t, result.Context)
}

func TestGenerateObjectWithProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := ObjectParams[testRecipe]{
		Prompt:   "Generate a recipe",
		Provider: mockProvider,
		Schema:   testRecipeSchema(),
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})
	mockProvider.EXPECT().GenerateText(gomock.Any(), gomock.Any()).Return(models.LanguageModelOutput{
		Text:      `{"name":"soup","servings":2}`,
		ModelName: modelName,
	}, nil)

	result, err := GenerateObject(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, "soup", result.Object.Name)
	assert.Equal(t, 2, result.Object.Servings)
}

func TestGenerateObjectRepairLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Generate a recipe"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := ObjectParams[testRecipe]{
		Prompt:  prompt,
		Provider: mockProvider,
		Schema:  testRecipeSchema(),
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})

	// First attempt: invalid JSON.
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
			return len(p.Messages) == 1
		}),
	).Return(models.LanguageModelOutput{
		Text:      `not json`,
		ModelName: modelName,
		Usage:     models.LanguageModelUsage{OutputTokens: 10, InputTokens: 20, TotalTokens: 30},
	}, nil)

	// Second attempt: the repair messages are present, then valid JSON.
	mockProvider.EXPECT().GenerateText(
		gomock.Any(),
		gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
			if len(p.Messages) != 3 {
				return false
			}
			// assistant message with the bad output
			if p.Messages[1].Content().Text() != `not json` {
				return false
			}
			// human repair instruction
			return p.Messages[2].Type() == models.MessageRoleHuman
		}),
	).Return(models.LanguageModelOutput{
		Text:      `{"name":"fixed","servings":1}`,
		ModelName: modelName,
		Usage:     models.LanguageModelUsage{OutputTokens: 5, InputTokens: 10, TotalTokens: 15},
	}, nil)

	result, err := GenerateObject(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, "fixed", result.Object.Name)
	assert.Equal(t, 1, result.Object.Servings)
	// Usage from both attempts is accumulated.
	assert.Equal(t, int64(30), result.Usage.InputTokens)
	assert.Equal(t, int64(15), result.Usage.OutputTokens)
	assert.Equal(t, int64(45), result.Usage.TotalTokens)
}

func TestGenerateObjectExhaustedRetries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Generate a recipe"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := ObjectParams[testRecipe]{
		Prompt:          prompt,
		Provider:        mockProvider,
		Schema:          testRecipeSchema(),
		MaxObjectRetries: 2,
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})
	mockProvider.EXPECT().GenerateText(gomock.Any(), gomock.Any()).Return(models.LanguageModelOutput{
		Text:      `not json`,
		ModelName: modelName,
	}, nil).Times(2)

	_, err := GenerateObject(ctx, params)
	require.Error(t, err)

	var noObjectErr *models.NoObjectGeneratedError
	require.ErrorAs(t, err, &noObjectErr)
	assert.Equal(t, `not json`, noObjectErr.Text)

	var parseErr *models.ObjectParseError
	require.ErrorAs(t, err, &parseErr)
}

func TestGenerateObjectSchemaRejection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prompt := "Generate a recipe"
	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := ObjectParams[testRecipe]{
		Prompt:          prompt,
		Provider:        mockProvider,
		Schema:          testRecipeSchema(),
		MaxObjectRetries: 1,
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{
		ModelName: modelName,
	})
	// Valid JSON, but the validator rejects it (empty name).
	mockProvider.EXPECT().GenerateText(gomock.Any(), gomock.Any()).Return(models.LanguageModelOutput{
		Text:      `{"name":"","servings":4}`,
		ModelName: modelName,
	}, nil).Times(1)

	_, err := GenerateObject(ctx, params)
	require.Error(t, err)

	var noObjectErr *models.NoObjectGeneratedError
	require.ErrorAs(t, err, &noObjectErr)

	var validationErr *models.ObjectValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "recipe", validationErr.SchemaName)
}
