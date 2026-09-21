package graph

import (
	"context"
	"fmt"
)

type nodeResult[T any] struct {
	id    ID
	state T
	err   error
}

type NodeFn[T any] = func(ctx context.Context, state T) (T, error)

type stateNode[T any] struct {
	ID ID
	fn NodeFn[T]
}

func (n stateNode[T]) execute(ctx context.Context, state T, target Target) (delta T, err error) {
	threadId := ctx.Value("threadId").(ID)

	defer func() {
		if r := recover(); r != nil {
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

func (n stateNode[T]) id() ID {
	return n.ID
}

type WorkerNodeFn[T any] = func(ctx context.Context, params any) (T, error)

type workerNode[T any] struct {
	ID ID
	fn WorkerNodeFn[T]
}

func (n workerNode[T]) execute(ctx context.Context, state T, target Target) (delta T, err error) {
	threadId := ctx.Value("threadId").(ID)
	defer func() {
		if r := recover(); r != nil {
			err = NewNodeExecutionError(threadId, n.ID, fmt.Errorf("panic in node execution: %v", r))
		}
	}()

	return n.fn(ctx, target.params)
}

func (n workerNode[T]) id() ID {
	return n.ID
}

type node[T any] interface {
	execute(ctx context.Context, state T, target Target) (delta T, err error)
	id() ID
}
