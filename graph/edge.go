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
type NamedRouter[T any] = func(state T) string
type NamedRouterMap = map[string][]Target

type stateEdge[T any, D any] struct {
	start  ID
	end    []ID
	endMap NamedRouterMap
	router Router[T]
}

func (e stateEdge[T, D]) route(threadId ID, state T) (target []Target, err *RouterExecutionError) {
	defer func() {
		if r := recover(); r != nil {
			err = NewRouterExecutionError(threadId, e.start, fmt.Errorf("panic in router execution: %v", r))
		}
	}()

	if e.router != nil {
		// Expect the router to handle errors internally and return a valid list
		// of next nodes, we handle panics here to avoid crashing the entire
		// graph execution
		return e.router(state), nil
	}

	return IDs(e.end...), nil
}
