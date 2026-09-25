package graph

type InMemoryCheckpointer[T any, D any] struct {
	states  map[string]T
	steps   map[string][]ID
	results map[string]map[ID]D
}

func NewInMemoryCheckpointer[T any, D any]() *InMemoryCheckpointer[T, D] {
	return &InMemoryCheckpointer[T, D]{
		states:  make(map[string]T),
		steps:   make(map[string][]ID),
		results: make(map[string]map[ID]D),
	}
}

func (c *InMemoryCheckpointer[T, D]) Checkpoint(threadId string, cp Checkpoint[T, D]) error {
	logger.Debug("Checkpoint thread", "threadId", threadId, "state", cp.State, "steps", cp.Steps)
	c.states[threadId] = cp.State
	c.steps[threadId] = cp.Steps
	c.results[threadId] = cp.Results
	return nil
}

func (c *InMemoryCheckpointer[T, D]) Restore(threadId string) (Checkpoint[T, D], error) {
	logger.Debug("Restoring checkpoint", "threadId", threadId)
	if _, exists := c.states[threadId]; !exists {
		return Checkpoint[T, D]{}, NewCheckpointNotFoundError(threadId)
	}

	return Checkpoint[T, D]{
		State:   c.states[threadId],
		Steps:   c.steps[threadId],
		Results: c.results[threadId],
	}, nil
}
