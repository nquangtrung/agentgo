package utils

import (
	"time"
)

type Action[T any, A any] func(iteration int, acc *A) (T, error)
type ShouldContinueFn[T any, A any] func(iteration int, acc *A) bool
type Accumulator[T any, A any] func(acc *A, item T)

type RunnerEventType string

const (
	RunnerEventStart   RunnerEventType = "START"
	RunnerEventSuccess RunnerEventType = "SUCCESS"
	RunnerEventError   RunnerEventType = "ERROR"
)

type RunnerEvent[T any, A any] struct {
	Type       RunnerEventType
	ActionName string
	Iteration  int
	Result     T
	Error      error
	Timestamp  time.Time
	Acc        *A
}

type Runner[T any, A any] struct {
	EventChannel chan RunnerEvent[T, A]
	Accumulator  Accumulator[T, A]
}

func (r *Runner[T, A]) Execute(name string, action Action[T, A], iteration int, acc *A) error {
	if r.EventChannel != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       RunnerEventStart,
			ActionName: name,
			Iteration:  iteration,
			Acc:        acc,
		}
	}

	result, err := action(iteration, acc)
	if err != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       RunnerEventError,
			ActionName: name,
			Iteration:  iteration,
			Acc:        acc,
		}
		return err
	}

	if r.EventChannel != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       RunnerEventSuccess,
			ActionName: name,
			Iteration:  iteration,
			Acc:        acc,
			Result:     result,
		}
	}

	if r.Accumulator != nil {
		r.Accumulator(acc, result)
	}

	return nil
}

func (r *Runner[T, A]) Loop(
	name string,
	shouldContinueFn ShouldContinueFn[T, A],
	action Action[T, A],
	acc *A,
	breakIfError bool,
) {
	for iteration := 0; shouldContinueFn(iteration, acc); iteration++ {
		err := r.Execute(name, action, iteration, acc)
		if err != nil && breakIfError {
			break
		}
	}
}
