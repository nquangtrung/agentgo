package graph

import "fmt"

type Checkpoint[T any, D any] struct {
	State   T
	Steps   []ID
	Results map[ID]D
}

type Checkpointer[T any, D any] interface {
	Checkpoint(threadId string, cp Checkpoint[T, D]) error
	Restore(threadId string) (Checkpoint[T, D], error)
}

type CheckPointNotFoundError struct {
	ThreadId string
}

func (e *CheckPointNotFoundError) Error() string {
	return fmt.Sprintf("Checkpoint not found for threadId: %s", e.ThreadId)
}

func NewCheckpointNotFoundError(threadId string) *CheckPointNotFoundError {
	return &CheckPointNotFoundError{ThreadId: threadId}
}
