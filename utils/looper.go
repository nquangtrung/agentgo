package utils

import (
	"log"
	"time"
)

type LoopAction[T any, A any] func(iteration int, acc *A) (T, bool, error)
type Action[T any, A any] func(acc *A) (T, error)
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

func (r *Runner[T, A]) executeWithKeepGoing(name string, action LoopAction[T, A], iteration int, acc *A) (bool, error) {
	if r.EventChannel != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       RunnerEventStart,
			ActionName: name,
			Iteration:  iteration,
			Acc:        acc,
		}
	}

	result, keepGoing, err := action(iteration, acc)
	if err != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       RunnerEventError,
			ActionName: name,
			Iteration:  iteration,
			Acc:        acc,
		}
		return keepGoing, err
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

	return keepGoing, nil
}

func (r *Runner[T, A]) Execute(name string, action Action[T, A], acc *A) error {
	if r.EventChannel != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       RunnerEventStart,
			ActionName: name,
			Acc:        acc,
		}
	}

	result, err := action(acc)
	if err != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       RunnerEventError,
			ActionName: name,
			Acc:        acc,
		}
		return err
	}

	if r.EventChannel != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       RunnerEventSuccess,
			ActionName: name,
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
	action LoopAction[T, A],
	acc *A,
	breakIfError bool,
) {
	for iteration := 0; shouldContinueFn(iteration, acc); iteration++ {
		keepGoing, err := r.executeWithKeepGoing(name, action, iteration, acc)
		if !keepGoing {
			log.Printf("Loop %s: breaking loop at iteration %d due to keepGoing=false", name, iteration)
			break
		}
		if err != nil && breakIfError {
			log.Printf("Loop %s: breaking loop at iteration %d due to error: %v", name, iteration, err)
			break
		}
	}
}

func NewRunner[T any, A any](channel chan RunnerEvent[T, A], accumulator Accumulator[T, A]) *Runner[T, A] {
	return &Runner[T, A]{
		EventChannel: channel,
		Accumulator:  accumulator,
	}
}
