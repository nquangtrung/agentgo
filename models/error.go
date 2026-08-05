package models

import "fmt"

type UnsupportedModelError struct {
	ModelName string
}

func (e *UnsupportedModelError) Error() string {
	return fmt.Sprintf("unsupported model: %s", e.ModelName)
}
