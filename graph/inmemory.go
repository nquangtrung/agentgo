package graph

type InMemoryCheckpointer[T any] struct {
	states map[string]T
	steps  map[string][]ID
}

func NewInMemoryCheckpointer[T any]() *InMemoryCheckpointer[T] {
	return &InMemoryCheckpointer[T]{
		states: make(map[string]T),
		steps:  make(map[string][]ID),
	}
}

func (c *InMemoryCheckpointer[T]) Checkpoint(threadId string, cp Checkpoint[T]) error {
	logger.Debug("Checkpoint thread", "threadId", threadId, "state", cp.State, "steps", cp.Steps)
	c.states[threadId] = cp.State
	c.steps[threadId] = cp.Steps
	return nil
}

func (c *InMemoryCheckpointer[T]) Restore(threadId string) (Checkpoint[T], error) {
	logger.Debug("Restoring checkpoint", "threadId", threadId)
	if _, exists := c.states[threadId]; !exists {
		return Checkpoint[T]{
			State: *new(T),
			Steps: []ID{},
		}, nil
	}

	return Checkpoint[T]{
		State: c.states[threadId],
		Steps: c.steps[threadId],
	}, nil
}
