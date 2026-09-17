package graph

import "fmt"

type Reducer[T any] = func(state1 T, state2 T) T

func executeReducer[T any](reducer Reducer[T], state1 T, state2 T) (newState T, err *ReducerExecutionError) {
	defer func() {
		if r := recover(); r != nil {
			err = NewReducerExecutionError(fmt.Errorf("panic in reducer execution: %v", r))
		}
	}()

	return reducer(state1, state2), nil
}
