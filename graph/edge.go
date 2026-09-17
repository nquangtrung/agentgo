package graph

import "fmt"

type Router[T any] = func(state T) []ID
type StateEdge[T any] struct {
	Start  ID
	End    []ID
	Router Router[T]
}

func (e StateEdge[T]) route(state T) (ids []ID, err *RouterExecutionError) {
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
	return e.End, nil
}
