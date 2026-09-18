package graph

import "fmt"

type Target struct {
	id     ID
	send   bool
	params any
}

func IDs(ids ...ID) []Target {
	targets := make([]Target, len(ids))
	for i, id := range ids {
		targets[i] = Target{id: id, send: true}
	}
	return targets
}

func Send(id ID, params any) Target {
	return Target{id: id, send: true, params: params}
}

type Router[T any] = func(state T) []Target

type stateEdge[T any] struct {
	Start  ID
	End    []ID
	Router Router[T]
}

func (e stateEdge[T]) route(state T) (target []Target, err *RouterExecutionError) {
	defer func() {
		if r := recover(); r != nil {
			err = NewRouterExecutionError(e.Start, fmt.Errorf("panic in router execution: %v", r))
		}
	}()

	if e.Router != nil {
		// Expect the router to handle errors internally and return a valid list
		// of next nodes, we handle panics here to avoid crashing the entire
		// graph execution
		return e.Router(state), nil
	}

	return IDs(e.End...), nil
}
