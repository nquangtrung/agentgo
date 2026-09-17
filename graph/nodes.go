package graph

const (
	START ID = "start"
	END   ID = "end"
)

func bypassFn[T any](state T) T {
	return state
}

func newStartNode[T any]() StateNode[T] {
	return StateNode[T]{
		ID: START,
		fn: bypassFn[T],
	}
}

func newEndNode[T any]() StateNode[T] {
	return StateNode[T]{
		ID: END,
		fn: bypassFn[T],
	}
}
