package graph

const (
	START ID = "start"
	END   ID = "end"
)

func bypassFn[T any](state T) T {
	return state
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
