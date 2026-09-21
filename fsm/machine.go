package fsm

import (
	"context"
	"log/slog"
)

const TAG string = "[FSM]"

type State[T any] interface {
	Execute(ctx context.Context, fsmCtx *T) (State[T], error)
}

type FSM[T any] struct {
	currentState State[T]
}

func (fsm *FSM[T]) SetState(state State[T]) {
	fsm.currentState = state
}

func (fsm *FSM[T]) Run(ctx context.Context, initialState State[T], fsmCtx *T) error {
	logger.Info("Starting FSM with initial state", slog.Any("state", initialState))
	fsm.currentState = initialState

	for fsm.currentState != nil {
		logger.Info("Executing state", slog.Any("state", fsm.currentState))
		if err := ctx.Err(); err != nil {
			return ctx.Err()
		}
		nextState, err := fsm.currentState.Execute(ctx, fsmCtx)

		if err != nil {
			return err
		}
		fsm.currentState = nextState
	}
	return nil
}

func New[T any]() *FSM[T] {
	return &FSM[T]{
		currentState: nil,
	}
}
