package graph

import "fmt"

type nodeResult[T any] struct {
	id    ID
	state T
	err   *NodeExecutionError
}

type NodeFn[T any] = func(state T) T

type stateNode[T any] struct {
	ID ID
	fn NodeFn[T]
}

func (n stateNode[T]) execute(state T, target Target) (delta T, err *NodeExecutionError) {
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

func (n stateNode[T]) id() ID {
	return n.ID
}

type WorkerNodeFn[T any] = func(params any) T

type workerNode[T any] struct {
	ID ID
	fn WorkerNodeFn[T]
}

func (n workerNode[T]) execute(state T, target Target) (delta T, err *NodeExecutionError) {
	defer func() {
		if r := recover(); r != nil {
			err = NewNodeExecutionError(n.ID, fmt.Errorf("panic in node execution: %v", r))
		}
	}()

	return n.fn(target.params), nil
}

func (n workerNode[T]) id() ID {
	return n.ID
}

type node[T any] interface {
	execute(state T, target Target) (delta T, err *NodeExecutionError)
	id() ID
}
