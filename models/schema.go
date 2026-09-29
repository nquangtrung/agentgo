package models

import "encoding/json"

// ObjectSchema is the non-generic view of a schema, used by graph nodes and
// providers that cannot know the caller's concrete type.
type ObjectSchema interface {
	Name() string
	Description() string
	JSONSchema() map[string]any
	Validate(value any) error
	// DecodeJSON decodes raw JSON into the schema's concrete type.
	DecodeJSON(raw []byte) (any, error)
}

// Schema[T] describes the shape of an object the model must produce.
// It carries the JSON Schema sent to the provider plus a typed validator.
type Schema[T any] struct {
	name        string
	description string
	jsonSchema  map[string]any
	validate    func(T) error
}

type NewSchemaParams[T any] struct {
	Name        string
	Description string
	JSONSchema  map[string]any
	Validate    func(T) error
}

func NewSchema[T any](params NewSchemaParams[T]) Schema[T] {
	return Schema[T]{
		name:        params.Name,
		description: params.Description,
		jsonSchema:  params.JSONSchema,
		validate:    params.Validate,
	}
}

func (s Schema[T]) Name() string {
	return s.name
}

func (s Schema[T]) Description() string {
	return s.description
}

func (s Schema[T]) JSONSchema() map[string]any {
	return s.jsonSchema
}

// Validate type-asserts the decoded value back to T and runs the typed
// validator. A nil validator accepts any value of the right type.
func (s Schema[T]) Validate(value any) error {
	typed, ok := value.(T)
	if !ok {
		return &ObjectValidationError{
			SchemaName: s.name,
			Err:        &typeMismatchError{expected: s.name},
		}
	}
	if s.validate == nil {
		return nil
	}
	if err := s.validate(typed); err != nil {
		return &ObjectValidationError{SchemaName: s.name, Err: err}
	}
	return nil
}

// DecodeJSON unmarshals raw JSON into T. The graph works with any, so this is
// the bridge that produces the concrete type Validate expects.
func (s Schema[T]) DecodeJSON(raw []byte) (any, error) {
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

// typeMismatchError is returned when a decoded value does not match the
// schema's concrete type.
type typeMismatchError struct {
	expected string
}

func (e *typeMismatchError) Error() string {
	return "value does not match schema type: " + e.expected
}
