package models

import "fmt"

type UnsupportedModelError struct {
	ModelName string
}

func (e *UnsupportedModelError) Error() string {
	return fmt.Sprintf("unsupported model: %s", e.ModelName)
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
