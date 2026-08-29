package utils

import (
	"time"
)

type Action[T any, A any] func(acc *A) (T, error)
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
	Result     T
	Error      error
	Timestamp  time.Time
	Acc        *A
}

type Runner[T any, A any] struct {
	EventChannel chan RunnerEvent[T, A]
	Accumulator  Accumulator[T, A]
}

func (r *Runner[T, A]) EmitStart(event RunnerEventType, name string, acc *A) {
	if r.EventChannel != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       event,
			ActionName: name,
			Timestamp:  time.Now(),
			Acc:        acc,
		}
	}
}

func (r *Runner[T, A]) EmitSuccess(event RunnerEventType, name string, result T, acc *A) {
	if r.EventChannel != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       event,
			ActionName: name,
			Result:     result,
			Timestamp:  time.Now(),
			Acc:        acc,
		}
	}
}

func (r *Runner[T, A]) EmitError(event RunnerEventType, name string, err error, acc *A) {
	if r.EventChannel != nil {
		r.EventChannel <- RunnerEvent[T, A]{
			Type:       event,
			ActionName: name,
			Error:      err,
			Timestamp:  time.Now(),
			Acc:        acc,
		}
	}
}

func (r *Runner[T, A]) Execute(name string, action Action[T, A], acc *A) error {
	r.EmitStart(RunnerEventStart, name, acc)

	result, err := action(acc)
	if err != nil {
		r.EmitError(RunnerEventError, name, err, acc)
		return err
	}

	r.EmitSuccess(RunnerEventSuccess, name, result, acc)

	if r.Accumulator != nil {
		r.Accumulator(acc, result)
	}

	return nil
}

func (r *Runner[T, A]) Close() {
	if r.EventChannel != nil {
		close(r.EventChannel)
	}
}

func (r *Runner[T, A]) Channel() <-chan RunnerEvent[T, A] {
	return r.EventChannel
}

func NewRunner[T any, A any](accumulator Accumulator[T, A]) *Runner[T, A] {
	return &Runner[T, A]{
		EventChannel: make(chan RunnerEvent[T, A]),
		Accumulator:  accumulator,
	}
}

func NewRunnerNoEmit[T any, A any](accumulator Accumulator[T, A]) *Runner[T, A] {
	return &Runner[T, A]{
		EventChannel: nil,
		Accumulator:  accumulator,
	}
}
