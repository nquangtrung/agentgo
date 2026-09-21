package graph

import "context"

const (
	START ID = "start"
	END   ID = "end"
)

func bypassFn[T any](_ context.Context, state T) (T, error) {
	return state, nil
}

func newStartNode[T any]() stateNode[T] {
	return stateNode[T]{
		ID: START,
		fn: bypassFn[T],
	}
}

func newEndNode[T any]() stateNode[T] {
	return stateNode[T]{
		ID: END,
		fn: bypassFn[T],
	}
}
