package models

// ResponseFormat tells a provider how the model's output should be shaped.
// A nil ResponseFormat means plain text.
type ResponseFormat struct {
	Name        string
	Description string
	JSONSchema  map[string]any
	Strict      bool
}
