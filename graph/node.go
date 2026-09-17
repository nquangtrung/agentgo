package graph

import "fmt"

type nodeResult[T any] struct {
	id    ID
	state T
	err   *NodeExecutionError
}

type NodeFn[T any] = func(state T) T

type StateNode[T any] struct {
	ID ID
	fn NodeFn[T]
}

func (n StateNode[T]) execute(state T) (newState T, err *NodeExecutionError) {
	defer func() {
		if r := recover(); r != nil {
			err = NewNodeExecutionError(n.ID, fmt.Errorf("panic in node execution: %v", r))
		}
	}()

	// expect the node to handle errors internally and return a
	// valid state, we handle panics here to avoid crashing the
	// entire graph execution
	return n.fn(state), nil
}
