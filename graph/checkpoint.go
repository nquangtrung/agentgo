package graph

type Checkpoint[T any] struct {
	State T
	Steps []ID
}

type Checkpointer[T any] interface {
	Checkpoint(threadId string, cp Checkpoint[T]) error
	Restore(threadId string) (Checkpoint[T], error)
}
