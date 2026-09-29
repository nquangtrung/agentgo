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

func TestStreamObject(t *testing.T) {
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
	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
			if len(p.Messages) != 1 || p.Messages[0].Content().Text() != prompt {
				return false
			}
			return p.ResponseFormat != nil && p.ResponseFormat.Name == "recipe"
		}),
		gomock.Any(),
	).Do(func(ctx context.Context, p providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		context := models.LanguageModelContext{ModelName: modelName}
		emitter.Emit(models.NewTextPart(context, `{"name":"panc`))
		emitter.Emit(models.NewTextPart(context, `akes","servings":4}`))
	}).Return(models.LanguageModelOutput{
		Text: `{"name":"pancakes","servings":4}`,
		Usage: models.LanguageModelUsage{
			OutputTokens: 123,
			InputTokens:  245,
			TotalTokens:   368,
		},
		ModelName: modelName,
	}, nil)

	output := StreamObject(ctx, params)
	require.NotNil(t, output.Channel)
	assert.Equal(t, modelName, output.ModelName)

	var texts []string
	var objectPart models.ObjectPart

	for part := range output.Channel {
		switch p := part.(type) {
		case models.ProcessStartPart:
		case models.StepStartPart:
		case models.TextPart:
			texts = append(texts, p.Text())
		case models.ObjectPart:
			objectPart = p
		case models.StepEndPart:
		case models.ProcessEndPart:
		default:
			t.Fatalf("unexpected part type: %s", part.Type())
		}
	}

	assert.Equal(t, []string{`{"name":"panc`, `akes","servings":4}`}, texts)
	require.NotNil(t, objectPart, "should emit a final ObjectPart")
	assert.Equal(t, `{"name":"pancakes","servings":4}`, objectPart.Raw())

	recipe, ok := objectPart.Object().(testRecipe)
	require.True(t, ok, "ObjectPart should carry the schema's concrete type")
	assert.Equal(t, "pancakes", recipe.Name)
	assert.Equal(t, 4, recipe.Servings)
}

func TestStreamObjectError(t *testing.T) {
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
	mockProvider.EXPECT().StreamText(gomock.Any(), gomock.Any(), gomock.Any()).Do(
		func(ctx context.Context, p providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
			emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, `not json`))
		},
	).Return(models.LanguageModelOutput{
		Text:      `not json`,
		ModelName: modelName,
	}, nil)

	output := StreamObject(ctx, params)

	var errorPart models.ObjectErrorPart
	for part := range output.Channel {
		if p, ok := part.(models.ObjectErrorPart); ok {
			errorPart = p
		}
	}

	require.NotNil(t, errorPart, "should emit an ObjectErrorPart")
	assert.Equal(t, `not json`, errorPart.Raw())

	var parseErr *models.ObjectParseError
	require.ErrorAs(t, errorPart.Error(), &parseErr)
}

func TestStreamObjectRepairLoop(t *testing.T) {
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
	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
			return len(p.Messages) == 1
		}),
		gomock.Any(),
	).Do(func(ctx context.Context, p providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, `not json`))
	}).Return(models.LanguageModelOutput{Text: `not json`, ModelName: modelName}, nil)

	// Second attempt: repair messages present, then valid JSON.
	mockProvider.EXPECT().StreamText(
		gomock.Any(),
		gomock.Cond(func(p providers.AgentProviderPromptMessageParams) bool {
			return len(p.Messages) == 3
		}),
		gomock.Any(),
	).Do(func(ctx context.Context, p providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
		emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, `{"name":"fixed","servings":1}`))
	}).Return(models.LanguageModelOutput{
		Text:      `{"name":"fixed","servings":1}`,
		ModelName: modelName,
	}, nil)

	output := StreamObject(ctx, params)

	var objectPart models.ObjectPart
	for part := range output.Channel {
		if p, ok := part.(models.ObjectPart); ok {
			objectPart = p
		}
	}

	require.NotNil(t, objectPart, "should emit a final ObjectPart after repair")
	recipe, ok := objectPart.Object().(testRecipe)
	require.True(t, ok)
	assert.Equal(t, "fixed", recipe.Name)
}

func TestStreamObjectPartSequence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	modelName := "mocked-llm-3.6-flash"
	ctx := context.Background()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	params := ObjectParams[testRecipe]{
		Prompt:  "Generate a recipe",
		Provider: mockProvider,
		Schema:  testRecipeSchema(),
	}

	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: modelName})
	mockProvider.EXPECT().StreamText(gomock.Any(), gomock.Any(), gomock.Any()).Do(
		func(ctx context.Context, p providers.AgentProviderPromptMessageParams, emitter models.PartEmitter) {
			emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, `{"name":"a`))
			emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, `bc","servings":`))
			emitter.Emit(models.NewTextPart(models.LanguageModelContext{ModelName: modelName}, `2}`))
		},
	).Return(models.LanguageModelOutput{
		Text:      `{"name":"abc","servings":2}`,
		ModelName: modelName,
	}, nil)

	output := StreamObject(ctx, params)

	var types []models.PartType
	for part := range output.Channel {
		types = append(types, part.Type())
	}

	assert.Equal(t, []models.PartType{
		models.PartTypeStart,
		models.PartTypeStepStart,
		models.PartTypeText,
		models.PartTypeText,
		models.PartTypeText,
		models.PartTypeObject,
		models.PartTypeStepEnd,
		models.PartTypeEnd,
	}, types)
}
