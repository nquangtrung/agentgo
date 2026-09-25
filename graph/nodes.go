package graph

import "context"

const (
	START ID = "start"
	END   ID = "end"
)

func bypassFn[T any, D any](_ context.Context, state T) (D, error) {
	return *new(D), nil
}

func newStartNode[T any, D any]() stateNode[T, D] {
	return stateNode[T, D]{
		ID: START,
		fn: bypassFn[T, D],
	}
}

func newEndNode[T any, D any]() stateNode[T, D] {
	return stateNode[T, D]{
		ID: END,
		fn: bypassFn[T, D],
	}
}
