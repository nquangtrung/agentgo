package graph

import (
	"context"
	"fmt"
	"runtime/debug"
)

type nodeResult[T any, D any] struct {
	id    ID
	delta D
	err   error
}

type NodeFn[T any, D any] = func(ctx context.Context, state T) (D, error)

type stateNode[T any, D any] struct {
	ID ID
	fn NodeFn[T, D]
}

func (n stateNode[T, D]) execute(ctx context.Context, state T, target Target) (delta D, err error) {
	threadId := ctx.Value("threadId").(ID)

	defer func() {
		if r := recover(); r != nil {
			logger.Error(string(debug.Stack()))
			err = NewNodeExecutionError(threadId, n.ID, fmt.Errorf("panic in node execution: %v", r))
		}
	}()

	// expect the node to handle errors internally and return a
	// valid state, we handle panics here to avoid crashing the
	// entire graph execution
	result, err := n.fn(ctx, state)
	if err == nil {
		return result, nil
	}

	if e, ok := err.(withID); ok {
		e.SetID(n.ID)
	}
	if e, ok := err.(withThreadID); ok {
		e.SetThreadID(threadId)
	}

	return result, err
}

func (n stateNode[T, D]) id() ID {
	return n.ID
}

type WorkerNodeFn[T any, D any] = func(ctx context.Context, params any) (D, error)

type workerNode[T any, D any] struct {
	ID ID
	fn WorkerNodeFn[T, D]
}

func (n workerNode[T, D]) execute(ctx context.Context, state T, target Target) (delta D, err error) {
	threadId := ctx.Value("threadId").(ID)
	defer func() {
		if r := recover(); r != nil {
			logger.Error(string(debug.Stack()))
			err = NewNodeExecutionError(threadId, n.ID, fmt.Errorf("panic in node execution: %v", r))
		}
	}()

	return n.fn(ctx, target.params)
}

func (n workerNode[T, D]) id() ID {
	return n.ID
}

type node[T any, D any] interface {
	execute(ctx context.Context, state T, target Target) (delta D, err error)
	id() ID
}
