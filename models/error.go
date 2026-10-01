package models

import "fmt"

type UnsupportedModelError struct {
	ModelName string
}

func (e *UnsupportedModelError) Error() string {
	return fmt.Sprintf("unsupported model: %s", e.ModelName)
}

type ExecutionContextError struct {
	Message string
}

func (e *ExecutionContextError) Error() string {
	return fmt.Sprintf("execution context error: %s", e.Message)
}

type ToolExecutionError struct {
	ToolName string
	Err      error
}

func (e *ToolExecutionError) Error() string {
	return fmt.Sprintf("error executing tool %s: %v", e.ToolName, e.Err)
}

type ToolNotFoundError struct {
	ToolName string
}

func (e *ToolNotFoundError) Error() string {
	return fmt.Sprintf("tool not found: %s", e.ToolName)
}

// StructuredOutputUnsupportedError is returned when a caller requests schema-
// constrained output from a provider that cannot enforce it. Providers raise
// this instead of silently dropping the response format, which would otherwise
// surface later as a confusing parse error inside the object-repair loop.
type StructuredOutputUnsupportedError struct {
	ProviderName string
	Reason       string
}

func (e *StructuredOutputUnsupportedError) Error() string {
	return fmt.Sprintf("structured output is not supported by provider %s: %s", e.ProviderName, e.Reason)
}

// ObjectValidationError is returned when a decoded value does not satisfy the
// schema's validator.
type ObjectValidationError struct {
	SchemaName string
	Err        error
}

func (e *ObjectValidationError) Error() string {
	return fmt.Sprintf("object validation error for schema %s: %v", e.SchemaName, e.Err)
}

func (e *ObjectValidationError) Unwrap() error {
	return e.Err
}

// ObjectParseError is returned when the model's output is not valid JSON.
type ObjectParseError struct {
	Text string
	Err  error
}

func (e *ObjectParseError) Error() string {
	return fmt.Sprintf("object parse error: %v", e.Err)
}

func (e *ObjectParseError) Unwrap() error {
	return e.Err
}

// NoObjectGeneratedError is the top-level error returned when object generation
// fails after all retries are exhausted. It carries the raw text, the usage
// accumulated across attempts, and the underlying cause.
type NoObjectGeneratedError struct {
	Message   string
	Text      string
	Usage     LanguageModelUsage
	ModelName string
	Cause     error
}

func (e *NoObjectGeneratedError) Error() string {
	return fmt.Sprintf("no object generated: %s", e.Message)
}

func (e *NoObjectGeneratedError) Unwrap() error {
	return e.Cause
}
