package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recipe struct {
	Name     string `json:"name"`
	Servings int    `json:"servings"`
}

func recipeSchema() Schema[recipe] {
	return NewSchema[recipe](NewSchemaParams[recipe]{
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
		Validate: func(r recipe) error {
			if r.Name == "" {
				return errors.New("name must not be empty")
			}
			if r.Servings <= 0 {
				return errors.New("servings must be positive")
			}
			return nil
		},
	})
}

func TestSchemaImplementsObjectSchema(t *testing.T) {
	var _ ObjectSchema = recipeSchema()
}

func TestSchemaDecodeJSON(t *testing.T) {
	schema := recipeSchema()

	value, err := schema.DecodeJSON([]byte(`{"name":"pancakes","servings":4}`))
	require.NoError(t, err)

	typed, ok := value.(recipe)
	require.True(t, ok, "decoded value should be the schema's concrete type")
	assert.Equal(t, "pancakes", typed.Name)
	assert.Equal(t, 4, typed.Servings)
}

func TestSchemaDecodeJSONInvalid(t *testing.T) {
	schema := recipeSchema()

	_, err := schema.DecodeJSON([]byte(`not json`))
	require.Error(t, err)
}

func TestSchemaValidatePasses(t *testing.T) {
	schema := recipeSchema()

	value, err := schema.DecodeJSON([]byte(`{"name":"pancakes","servings":4}`))
	require.NoError(t, err)
	assert.NoError(t, schema.Validate(value))
}

func TestSchemaValidateRejects(t *testing.T) {
	schema := recipeSchema()

	value, err := schema.DecodeJSON([]byte(`{"name":"","servings":4}`))
	require.NoError(t, err)

	err = schema.Validate(value)
	require.Error(t, err)
	assert.IsType(t, &ObjectValidationError{}, err)
	assert.Contains(t, err.Error(), "recipe")
}

func TestSchemaValidateTypeMismatch(t *testing.T) {
	schema := recipeSchema()

	// A map[string]any is what a non-generic caller would hand in; it is not
	// the schema's concrete type.
	err := schema.Validate(map[string]any{"name": "pancakes"})
	require.Error(t, err)
	assert.IsType(t, &ObjectValidationError{}, err)
}

func TestSchemaNilValidatorAccepts(t *testing.T) {
	schema := NewSchema[map[string]any](NewSchemaParams[map[string]any]{
		Name:       "any",
		JSONSchema: map[string]any{"type": "object"},
	})

	value, err := schema.DecodeJSON([]byte(`{"a":1}`))
	require.NoError(t, err)
	assert.NoError(t, schema.Validate(value))
}

func TestSchemaAccessors(t *testing.T) {
	schema := recipeSchema()
	assert.Equal(t, "recipe", schema.Name())
	assert.Equal(t, "A recipe", schema.Description())
	assert.Equal(t, "object", schema.JSONSchema()["type"])
}
