package graph

import "fmt"

type Reducer[T any, D any] = func(state T, delta D) T

func executeReducer[T any, D any](reducer Reducer[T, D], state T, delta D) (newState T, err *ReducerExecutionError) {
	defer func() {
		if r := recover(); r != nil {
			err = NewReducerExecutionError(fmt.Errorf("panic in reducer execution: %v", r))
		}
	}()

	return reducer(state, delta), nil
}
