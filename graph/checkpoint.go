package graph

import "fmt"

type Checkpoint[T any] struct {
	State   T
	Steps   []ID
	Results map[ID]T
}

type Checkpointer[T any] interface {
	Checkpoint(threadId string, cp Checkpoint[T]) error
	Restore(threadId string) (Checkpoint[T], error)
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
